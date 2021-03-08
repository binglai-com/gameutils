package mongo

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/binglai-com/gameutils"
	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type DataBase struct {
	sync.RWMutex
	freesess []*mongoclient
	DailInfo *mgo.DialInfo //url配置信息

	//	DBSession chan *mongoclient

	MaxNum int
	CurNum int

	asyncinchan  chan func() func() //进入数据库工作线程的通道
	asyncoutchan chan func()        //数据库工作线程结束的通道
}

type mongoclient struct {
	c      *mgo.Session
	active bool
	lastT  time.Time //放回会话池的时间
}

func GetUID() string { ///分配一个唯一ID
	return bson.NewObjectId().Hex()
}

func GetShortUID() string { ///分配一个唯一ID
	uid := string(bson.NewObjectId())
	nid := uid[:4] + uid[8:]
	return fmt.Sprintf(`%x`, nid)
}

func (database *DataBase) IndexTable(dbname string, colname string, indexname string, key []string, unique bool, dropDups bool) {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	ins, _ := col.Indexes()
	for _, v := range ins {
		if v.Name == indexname {
			return
		}
	}
	index := mgo.Index{
		Key:        key,
		Unique:     unique,
		DropDups:   dropDups,
		Background: true, // See notes.
		Sparse:     false,
		Name:       indexname,
	}
	err := col.EnsureIndex(index)
	if err != nil {
		filelog.ERROR("mongodb", fmt.Sprintf("creat index error:%s", err.Error()))
	}
}

func (database *DataBase) FindCount(dbname string, colname string, find interface{}, res *int) (err error) {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	*res, err = col.Find(find).Count()
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("FindCount error:%s", err.Error()))
	}
	return err
}

//查找修饰条件
type FindDecorate struct {
	Selector   interface{}
	Skip       int
	Limit      int
	SortFileds []string
}

//根据查找修饰条件筛选记录
func (database *DataBase) FindMany(dbname string, colname string, find interface{}, decorate FindDecorate, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	query := col.Find(find)
	if decorate.Selector != nil {
		query = query.Select(decorate.Selector)
	}

	if decorate.Skip > 0 {
		query = query.Skip(decorate.Skip)
	}
	if decorate.Limit > 0 {
		query = query.Limit(decorate.Limit)
	}
	if decorate.SortFileds != nil && len(decorate.SortFileds) > 0 {
		query = query.Sort(decorate.SortFileds...)
	}

	err := query.All(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("FindMany error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) FindAllSelector(dbname string, colname string, find interface{}, selector interface{}, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.Find(find).Select(selector).All(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("FindAllSelector error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) FindAllSelectorLimit(dbname string, colname string, find interface{}, selector interface{}, limit int, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.Find(find).Select(selector).Limit(limit).All(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("FindAllSelector error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) GetCollectionNames(dbname string) []string {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	names, err := db.CollectionNames()
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("GetCollectionNames error:%s", err.Error()))
	}
	return names
}

func (database *DataBase) FindAll(dbname string, colname string, find interface{}, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.Find(find).All(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("FindAll error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) FindIter(dbname string, colname string, find interface{}, decorate FindDecorate) *mgo.Iter {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	query := col.Find(find)
	if decorate.Selector != nil {
		query.Select(decorate.Selector)
	}

	if decorate.Skip > 0 {
		query.Skip(decorate.Skip)
	}

	if decorate.Limit > 0 {
		query.Limit(decorate.Limit)
	}

	if decorate.SortFileds != nil && len(decorate.SortFileds) > 0 {
		query.Sort(decorate.SortFileds...)
	}

	return query.Iter()
}

func (database *DataBase) FindIterSelect(dbname string, colname string, find interface{}, selects interface{}) *mgo.Iter {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	return col.Find(find).Select(selects).Iter()
}

func (database *DataBase) FindOne(dbname string, colname string, find interface{}, decorate FindDecorate, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	query := col.Find(find)

	if decorate.Selector != nil {
		query.Select(decorate.Selector)
	}

	if decorate.Skip > 0 {
		query.Skip(decorate.Skip)
	}

	if decorate.Limit > 0 {
		query.Limit(decorate.Limit)
	}

	if decorate.SortFileds != nil && len(decorate.SortFileds) > 0 {
		query.Sort(decorate.SortFileds...)
	}

	err := query.One(result)
	if err != nil {
		if err.Error() != "not found" {
			//			conn.active = false
			filelog.ERROR("mongodb", fmt.Sprintf("FindOne error:%s", err.Error()))
		}
	}
	return err
}

func (database *DataBase) FindOneSelect(dbname string, colname string, find interface{}, selector interface{}, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	err := conn.c.DB(dbname).C(colname).Find(find).Select(selector).One(result)
	if err != nil {
		if err.Error() != "not found" {
			//			conn.active = false
			filelog.ERROR("mongodb", fmt.Sprintf("FindOneSelect error:%s", err.Error()))
		}
	}
	return err
}

func (database *DataBase) FindId(dbname string, colname string, id interface{}, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.FindId(id).One(result)
	if err != nil {
		if err.Error() != "not found" {
			//			conn.active = false
			filelog.ERROR("mongodb", fmt.Sprintf("FindId error:%s", err.Error()))
		}
		return err
	}
	return err
}

func (database *DataBase) FindBySkipLimit(dbname string, colname string, find interface{}, result interface{}, skip int, limit int, sortFields ...string) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	var err error
	if len(sortFields) == 0 {
		err = col.Find(find).Skip(skip).Limit(limit).All(result)
	} else {
		err = col.Find(find).Skip(skip).Sort(sortFields...).Limit(limit).All(result)
	}
	if err != nil {
		if err.Error() != "not found" {
			//			conn.active = false
			filelog.ERROR("mongodb", fmt.Sprintf("FindBySkipLimit error:%s", err.Error()))
		}
		return err
	}
	return err
}

//根据字段名保存数据
func (database *DataBase) SetFields(dbname string, colname string, sets interface{}, fieldsname ...string) error {
	var setfields = fieldsname
	if len(fieldsname) == 0 {
		setfields = nil //默认全部保存
	}
	datas, copyerr := gameutils.CopyByFields(sets, setfields)
	if copyerr != nil {
		return copyerr
	}

	var _id = gameutils.GetValueByBsonTag(sets, "_id")
	if _id == nil {
		return nil, fmt.Errorf("SetFields bson tag _id not found.")
	}

	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	return col.Update(bson.M{"_id": _id}, bson.M{"$set": datas})
}

func (database *DataBase) Update1(dbname string, colname string, selector interface{}, update interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)

	err := col.Update(selector, update)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("Update1 error:%s", err.Error()))
		return err
	}

	return err
}

func (database *DataBase) RunCmd(dbname string, cmd interface{}, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	err := db.Run(cmd, result)
	if err != nil && err.Error() != "not found" {
		filelog.ERROR("mongodb", fmt.Sprintf("RunCmd error:%s", err.Error()))
		return err
	}
	return err
}

//获取数据库自增id
func (database *DataBase) GetAutoIncId(dbname string, colname string) (int64, error) {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	var result = struct {
		LastErrorObject struct {
			N               int
			UpdatedExisting bool
		}
		Value bson.M
		Ok    int
	}{}
	err := db.Run(
		bson.D{
			{"findAndModify", "autocounters"},
			{"query", bson.M{"_id": colname}},
			{"update", bson.M{"$inc": bson.M{"seqid": int64(1)}}},
			{"upsert", true},
			{"new", true},
		},
		&result)
	if err != nil {
		filelog.ERROR("mongodb", fmt.Sprintf("GetAutoIncId error:%s", err.Error()))
		return 0, err
	}

	if result.Ok != 1 { //执行失败
		filelog.ERROR("mongodb", "GetAutoIncId ret ", result)
		return 0, fmt.Errorf("GetAutoIncId ret %v", result)
	}

	var retid = result.Value["seqid"].(int64)

	return retid, nil
}

//upsert
func (database *DataBase) Upsert(dbname string, colname string, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	info, err := col.Upsert(selector, update)
	if err != nil {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("Upsert error:%s", err.Error()))
		return info, err
	}
	return info, err
}

func (database *DataBase) UpdateAll(dbname string, colname string, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	info, err := col.UpdateAll(selector, bson.M{"$set": update})
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("UpdateAll error:%s", err.Error()))
		return info, err
	}
	return info, err
}

func (database *DataBase) Updatebyid(dbname string, colname string, id interface{}, update interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.UpdateId(id, bson.M{"$set": update})
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("UpdateByid error:%s", err.Error()), " data:", update)
		return err
	}
	return err
}

func (database *DataBase) Delete(dbname string, colname string, id interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.RemoveId(id)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("Delete error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) DropCol(dbname string, colname string) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.DropCollection()
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("DropCol error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) DropDB(dbname string) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	err := db.DropDatabase()
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("DropDB error:%s", err.Error()))
	}
	return err
}

func (database *DataBase) Insert(dbname string, colname string, data interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.Insert(data)
	if err != nil {
		filelog.ERROR("mongodb", fmt.Sprintf("Insert %s %s error:%s", dbname, colname, err.Error()))
		//		conn.active = false
		return err
	}
	return nil
}

func (database *DataBase) Insert_IgnoreDupKey(dbname string, colname string, data interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)
	err := col.Insert(data)
	if err != nil {
		if strings.Index(err.Error(), "E11000") == -1 { //非E11000错误
			filelog.ERROR("mongodb", fmt.Sprintf("Insert %s %s error:%s", dbname, colname, err.Error()))
		}
		return err
	}
	return nil
}

//聚合操作 返回结果是文档数组
func (database *DataBase) AggregateAll(dbname, colname string, pipeline []bson.M, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)

	err := col.Pipe(pipeline).All(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("AggregateAll error:%s", err.Error()))
	}
	return err
}

//聚合操作 返回结果是单个文档
func (database *DataBase) AggregateOne(dbname, colname string, pipeline []bson.M, result interface{}) error {
	conn := database.getdbsession()
	defer database.freesession(conn)
	db := conn.c.DB(dbname)
	col := db.C(colname)

	err := col.Pipe(pipeline).One(result)
	if err != nil && err.Error() != "not found" {
		//		conn.active = false
		filelog.ERROR("mongodb", fmt.Sprintf("AggregateOne error:%s", err.Error(), "pipeline:", pipeline))
	}
	return err
}

func (database *DataBase) FindGroupCountOne(dbname string, colname string, find interface{}, group interface{}, sort interface{}, limit int, result interface{}) error {
	m := []bson.M{
		{"$match": find},
		{"$group": group}}
	if sort != nil {
		m = append(m, bson.M{"$sort": sort})
	}
	if limit > 0 {
		m = append(m, bson.M{"$limit": limit})
	}

	return database.AggregateOne(dbname, colname, m, result)
}

func (database *DataBase) FindGroupCount(dbname string, colname string, find interface{}, group interface{}, sort interface{}, limit int, result interface{}) error {
	m := []bson.M{
		{"$match": find},
		{"$group": group}}
	if sort != nil {
		m = append(m, bson.M{"$sort": sort})
	}
	if limit > 0 {
		m = append(m, bson.M{"$limit": limit})
	}

	return database.AggregateAll(dbname, colname, m, result)
}
func NewID() string {
	return string(bson.NewObjectId())
}

func (c *DataBase) getdbsession() *mongoclient {
judgefree:
	c.Lock()
	if len(c.freesess) == 0 { //已到达上限 休眠1ms后再次尝试
		c.Unlock()
		time.Sleep(1000)
		goto judgefree
	}

	//取出可用的sess

	var lastidx = len(c.freesess) - 1
	var ret = c.freesess[lastidx]
	c.freesess = c.freesess[:lastidx]
	c.Unlock()
	ret.active = true //被取走后变成激活状态
	return ret

	//	select {
	//	case d := <-c.DBSession:
	//		c.CurNum++
	//		return d
	//	default:
	//		if c.CurNum >= c.MaxNum {
	//			filelog.ERROR("mongodb", "GetDBSession db curnum > maxnum, cannot create connect.", c.MaxNum)
	//			return nil
	//		}
	//		sess, err := mgo.Dial(c.DBConn)
	//		if err != nil {
	//			filelog.ERROR("mongodb", fmt.Sprintf("connect db error:%s", err.Error()))
	//			return nil
	//		}
	//		c.CurNum++
	//		sess.SetCursorTimeout(0)
	//		sess.SetSocketTimeout(0)
	//		sess.SetSyncTimeout(0)
	//		d := new(mongoclient)
	//		d.c = sess
	//		d.active = true
	//		return d
	//	}
}

func (c *DataBase) freesession(sess *mongoclient) {
	if sess == nil {
		return
	}
	c.Lock()
	sess.lastT = time.Now()
	//	if !sess.active {
	//		sess.c.Close()
	//	} else {
	c.freesess = append(c.freesess, sess)
	//	}
	c.Unlock()
}

func (c *DataBase) Init(gconn string) bool {
	var dailinfo, parseerr = mgo.ParseURL(gconn)
	if parseerr != nil {
		filelog.ERROR("mongodb", fmt.Sprintf("parse db url error:%s", parseerr.Error()))
		return false
	}

	c.DailInfo = dailinfo

	if c.MaxNum <= 0 {
		c.MaxNum = 1024 //最大会话数默认1024个
	}

	c.freesess = make([]*mongoclient, 0, c.MaxNum)

	//初始化第一个会话
	originsess, err := mgo.Dial(gconn)
	if err != nil {
		filelog.ERROR("mongodb", fmt.Sprintf("init dbpool error:%s", err.Error()))
		return false
	}
	originsess.SetMode(mgo.Strong, true)

	var now = time.Now()
	for i := 1; i < c.MaxNum; i++ { //初始化其他会话
		var scp = originsess.Copy()
		c.freesess = append(c.freesess, &mongoclient{
			scp,
			false, //通过dial和copy初始化出来的sess当前并未缓存任何socket连接 因此active状态为false
			now})
	}

	c.freesess = append(c.freesess, &mongoclient{ //源sess放到最后 保证优先被取走
		originsess,
		false,
		now})

	c.asyncinchan = make(chan func() func(), 65535)
	c.asyncoutchan = make(chan func(), 65535)
	go c.asyncworkprocess() //数据库工作线程
	go c.checkidlesess()
	return true
}

var (
	mili_10          = 10 * time.Millisecond
	mili_100         = 100 * time.Millisecond
	sessidle_timeout = 30 * time.Second //会话的IDLE 超时时间
)

//检查激活状态下IDLE TimeOut的sess 调用refesh方法
func (c *DataBase) checkidlesess() {
	for {
		time.Sleep(time.Second * 15) //每15秒检查一次
		var freecnt = 0
		c.Lock()
		var now = time.Now()
		for i := len(c.freesess) - 1; i >= 0; i-- {
			var sess = c.freesess[i] //从后往前
			if sess.active {
				if now.Sub(sess.lastT) >= sessidle_timeout {
					sess.c.Refresh()    //将缓存的socket放入连接池
					sess.active = false //将激活状态置为false
					freecnt++
				}
			} else { //按照后进先出的原则 这里应该已经到达未激活区域  后面的不用再检查了
				break
			}
		}
		c.Unlock()
		filelog.INFO("mongo", "free idle sess cnt : ", freecnt, " totalcost:", time.Since(now), " addrs : ", c.DailInfo.Addrs)
	}
}

func (c *DataBase) asyncworkprocess() {
	for {
		var proc = <-c.asyncinchan //从进队列中取出一个待处理任务
		cb := proc()
		if cb != nil { //拥有回调方法
			//将回调函数放入出队列
			for i := 0; i < 1000; i++ {
				ok := false
				select {
				case c.asyncoutchan <- cb:
					ok = true
				default:
					time.Sleep(mili_10)
				}

				if ok {
					break
				}
			}
		}

	}
}

type Call struct {
	Proc  func(v ...interface{}) (interface{}, error)
	Args  []interface{}
	Reply interface{}
	Done  chan *Call // Strobes when call is complete.
	Err   error
}

//在数据库线程中同步执行一段数据库操作过程
func (c *DataBase) SyncProcess(proc func(v ...interface{}) (interface{}, error), params ...interface{}) (ret interface{}, reterror error) {
	if call, err := c.Process(proc, make(chan *Call, 1), params...); err != nil {
		reterror = err
		return
	} else {
		reply := <-call.Done
		ret, reterror = reply.Reply, reply.Err
		return
	}
}

//在数据库线程中执行一段代码
func (c *DataBase) Process(proc func(v ...interface{}) (interface{}, error), done chan *Call, params ...interface{}) (ret *Call, err error) {
	err = nil
	if done == nil {
		done = make(chan *Call, 10)
	} else {
		if cap(done) == 0 {
			err = errors.New("mongo Process : done channel is unbuffered ")
			return
		}
	}

	tmp := new(Call)
	tmp.Proc = proc
	tmp.Args = params
	tmp.Done = done

	asynccb := func() func() {
		tmp.Reply, tmp.Err = proc(params...)
		tmp.Done <- tmp
		return nil
	}

	select {
	case c.asyncinchan <- asynccb:
		ret = tmp
	default:
		err = errors.New("Process In Buff Full")
	}

	return
}

//根据字段名保存数据 saved 需要为结构体指针类型 saved内需要有bson标记为 "_id"的字段 否则保存失败
func (c *DataBase) SaveDataByFields(saved interface{}, savefileds []string, dbname string, colname string, bosync bool) (ret *Call, err error) {
	datas, copyerr := gameutils.CopyByFields(saved, savefileds)
	if copyerr != nil {
		return nil, copyerr
	}

	var _id = gameutils.GetValueByBsonTag(saved, "_id")
	if _id == nil {
		return nil, fmt.Errorf("SaveDataByFields bson tag _id not found.")
	}

	ret, err = c.Process(func(v ...interface{}) (interface{}, error) {
		reterr := c.Update1(dbname, colname, bson.M{"_id": _id}, bson.M{"$set": v[0]})
		if reterr != nil {
			filelog.ERROR("mongo", "SaveDataByFields update err : ", reterr.Error())
		}
		return nil, reterr
	}, nil, datas)

	if err != nil {
		return
	}

	if bosync {
		<-ret.Done
	}
	return
}

//异步执行数据库操作 防止阻塞当前工作线程
func (c *DataBase) AsyncProcess(proc func(v ...interface{}) (interface{}, error), cb func(interface{}, error), params ...interface{}) error {
	if proc == nil {
		return errors.New("AsyncProcess proc nil")
	}

	asynccb := func() func() {
		ret, err := proc(params...)
		if cb != nil {
			return func() {
				cb(ret, err)
			}
		} else {
			return nil
		}
	}

	select {
	case c.asyncinchan <- asynccb:
		return nil
	default:
		return errors.New("AsyncProcess In Buff Full")
	}

	return nil
}

//轮询异步回调
func (c *DataBase) UpdateAsyncCallback() {
	for i := 0; i < 1000; i++ {
		select {
		case cb := <-c.asyncoutchan:
			cb()
		default:
			return
		}
	}
}

//判断是否还有未完成的数据库任务
func (c *DataBase) CntUnFinishedWork() int {
	return len(c.asyncinchan)
}
