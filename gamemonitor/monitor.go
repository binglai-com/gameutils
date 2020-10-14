package gamemonitor

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/binglai-com/gameutils/gamedb/mongo"
	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/globalsign/mgo/bson"
	"github.com/shirou/gopsutil/process"
)

//core监控器
type monitor struct {
	WCoreStart
	*filelog.NewFileLog
	LogDB      *mongo.DataBase
	report     WReport
	mostTime   *list.List
	threadMap  map[string]*thread
	stamp      time.Time
	interval   time.Duration //统计间隔（秒）
	isRun      bool
	recentPeak int
	mostSize   *list.List
	mostDB     *list.List
}

const (
	monitor_throughput = 1 //吞吐量
	monitor_respond    = 2 //指令
	monitor_thread     = 3 //线程
	monitor_curonline  = 4 //上线人数
	monitor_dbcost     = 5 //db耗时
)

type monitorinfo struct {
	cmd  int
	data interface{}
}

//线程信息
type thread struct {
	name  string
	val   time.Duration
	max   time.Duration
	min   time.Duration
	total time.Duration
	tick  int
}

var g_monitor monitor
var LOG *filelog.NewFileLog
var monitor_info chan monitorinfo

//放入信息
func putMonitorInfo(cmd int, data interface{}) {
	if !g_monitor.isRun { //线程不在运行
		return
	}
	select {
	case monitor_info <- monitorinfo{cmd: cmd, data: data}:
	default:
		LOG.ERROR("chan", "channel monitor_info error buffer full")
	}
}

//获取所有信息
func getAllMonitorInfo() []monitorinfo {
	info := make([]monitorinfo, 0)
	for {
		select {
		case v := <-monitor_info:
			info = append(info, v)
		default:
			return info
		}
	}
}

//core监控(直接到数据库)
func CoreMonitor(info WCoreStart, log *filelog.NewFileLog) bool {
	if g_monitor.isRun == true {
		return false
	}
	monitor_info = make(chan monitorinfo, 65530)
	if !g_monitor.LogDB.Init(info.DBGLogConn) {
		return false
	}
	g_monitor.isRun = true
	LOG = log
	g_monitor.WCoreStart = info
	g_monitor.interval = 5 * time.Minute //5分钟汇报一次
	g_monitor.stamp = time.Now()
	g_monitor.mostTime = list.New()
	g_monitor.mostSize = list.New()
	g_monitor.mostDB = list.New()
	g_monitor.threadMap = make(map[string]*thread)
	go func() {
		defer func() {
			g_monitor.isRun = false
			if err := recover(); err != nil {
				LOG.ERROR("gocrash", "monitor error:", string(debug.Stack()))
			}
		}()
		for {
			start := time.Now()
			info := getAllMonitorInfo()
			for _, v := range info {
				switch v.cmd {
				case monitor_throughput:
					if m, ok := v.data.(WTopSize); ok {
						g_monitor.addThroughput(m)
					}
				case monitor_respond:
					if m, ok := v.data.(WRespond); ok {
						g_monitor.addRespond(m)
					}
				case monitor_thread:
					if m, ok := v.data.(thread); ok {
						g_monitor.addThread(m)
					}

				case monitor_curonline:
					if m, ok := v.data.(int); ok {
						g_monitor.addCurrentOnline(m)
					}
				case monitor_dbcost:
					if m, ok := v.data.(WTopDBCost); ok {
						g_monitor.addTopDBCost(m)
					}
				}

			}
			costTime := time.Since(start)
			SendThread("monitor", costTime)
			if time.Since(g_monitor.stamp) > g_monitor.interval { //5分钟到了汇报
				g_monitor.commit()
			}
			sleepT := 10*time.Millisecond - costTime
			if sleepT > time.Millisecond {
				time.Sleep(sleepT)
			}
		}
	}()
	LOG.INFO("monitor", "monitor start")
	return true
}

//上报web平台游戏上线
func PlatOnline(info WCoreStart, log *filelog.NewFileLog) []byte {
	if g_monitor.isRun == true {
		return []byte("fail")
	}
	monitor_info = make(chan monitorinfo, 65530)
	g_monitor.isRun = true
	LOG = log
	g_monitor.WCoreStart = info
	g_monitor.interval = 5 * time.Minute //5分钟汇报一次
	g_monitor.stamp = time.Now()
	g_monitor.mostTime = list.New()
	g_monitor.mostSize = list.New()
	g_monitor.mostDB = list.New()
	g_monitor.threadMap = make(map[string]*thread)
	go func() {
		defer func() {
			g_monitor.isRun = false
			if err := recover(); err != nil {
				LOG.ERROR("gocrash", "monitor error:", string(debug.Stack()))
			}
		}()
		for {
			start := time.Now()
			info := getAllMonitorInfo()
			//LOG.INFO("monitor","消息长度：", len(info))
			for _, v := range info {
				switch v.cmd {
				case monitor_throughput:
					if m, ok := v.data.(WTopSize); ok {
						g_monitor.addThroughput(m)
					}
				case monitor_respond:
					if m, ok := v.data.(WRespond); ok {
						g_monitor.addRespond(m)
					}
				case monitor_thread:
					if m, ok := v.data.(thread); ok {
						g_monitor.addThread(m)
					}

				case monitor_curonline:
					if m, ok := v.data.(int); ok {
						g_monitor.addCurrentOnline(m)
					}
				case monitor_dbcost:
					if m, ok := v.data.(WTopDBCost); ok {
						g_monitor.addTopDBCost(m)
					}
				}

			}
			costTime := time.Since(start)
			SendThread("monitor", costTime)
			if time.Since(g_monitor.stamp) > g_monitor.interval { //5分钟到了汇报
				g_monitor.commit()
			}
			sleepT := 10*time.Millisecond - costTime
			if sleepT > time.Millisecond {
				time.Sleep(sleepT)
			}
		}
	}()
	LOG.INFO("monitor", "monitor start")
	return WebReport(g_monitor.CollConn, WCORESTART, info)
}

//上报web平台游戏下线
func PlatOffline() {
	if g_monitor.mostTime != nil {
		g_monitor.commit()
		var sd WCoreStart
		sd.ID = g_monitor.ID
		WebReport(g_monitor.CollConn, WCOREEND, sd)
	}

}

//汇报吞吐量信息
/*
 参数：
	count: 数据包长度
	flag:0为入；1为出
*/

func SendPacketSize(code uint16, slen int, flag int) {
	putMonitorInfo(monitor_throughput, WTopSize{Code: int(code), Fsize: slen, Flag: flag})
}

//db花费时间
func SendDBCost(cost time.Duration, optype uint16, dbtype int, colname string) {
	putMonitorInfo(monitor_dbcost, WTopDBCost{Cost: int(cost / time.Millisecond), Optype: optype, Dbtype: dbtype, Record: colname})
}

//上报指令明细
/*
	参数:
		cmd:命令号
		data:数据
		cost:耗时(纳秒)
*/
func SendRespond(cmd uint16, data []byte, cost int64) {
	var m WRespond
	m.Cmd = cmd
	if data == nil {
		m.Data = ""
	} else {
		m.Data = string(data)
	}
	m.Cost = int(cost / int64(time.Millisecond))
	putMonitorInfo(monitor_respond, m)
}

//上报线程信息
/*
	参数:
		name:线程名称
		cost:耗时(纳秒)
*/
func SendThread(name string, cost time.Duration) {
	var m thread
	m.name = name
	m.val = cost
	if cost < time.Millisecond {
		return
	}
	putMonitorInfo(monitor_thread, m)
}

//上报在线人数
/*
	参数:
		total:当前在线人数
*/
func SendCurOnline(total int) {
	putMonitorInfo(monitor_curonline, total)
}

//添加吞吐量信息
func (this *monitor) addThroughput(m WTopSize) {
	if m.Flag == 0 {
		this.report.In += m.Fsize
	} else {
		this.report.Out += m.Fsize
	}
	if this.mostSize.Len() < 10 {
		this.addTopSize(m)
	} else {
		first := this.mostSize.Front()
		if first.Value.(WTopSize).Fsize < m.Fsize { //有大包
			this.addTopSize(m)
			this.mostSize.Remove(first)
		}
	}
}

//添加大包信息
func (this *monitor) addTopSize(m WTopSize) {
	flag := 0
	for elem := this.mostSize.Front(); elem != nil; elem = elem.Next() {
		cur := elem.Value.(WTopSize)
		if cur.Fsize > m.Fsize {
			this.mostSize.InsertBefore(m, elem)
			flag = 1
			break
		}
	}
	if flag == 0 {
		this.mostSize.PushBack(m)
	}
}

//添加指令耗时信息
func (this *monitor) addRespond(m WRespond) {
	this.report.Responds += 1
	flag := 0
	for elem := this.mostTime.Front(); elem != nil; elem = elem.Next() {
		cur := elem.Value.(WRespond)
		if cur.Cost > m.Cost {
			this.mostTime.InsertBefore(m, elem)
			flag = 1
			break
		}
	}
	if flag == 0 {
		this.mostTime.PushBack(m)
	}
	if this.report.Responds > 10 { //只记录10个
		this.mostTime.Remove(this.mostTime.Front())
	}
}

//添加线程信息
func (this *monitor) addThread(m thread) {
	if v, ok := this.threadMap[m.name]; ok {
		v.tick++
		if v.max < m.val {
			v.max = m.val
		} else if v.min > m.val {
			v.min = m.val
		}
		v.total += m.val
	} else { //新增线程信息
		m.tick++
		m.max = m.val
		m.min = m.val
		m.total = m.val
		this.threadMap[m.name] = &m
	}
}

func (this *monitor) addCurrentOnline(m int) {
	this.recentPeak = m
	if this.report.Peak < m {
		this.report.Peak = m
	}
}

func (this *monitor) addTopDBCost(m WTopDBCost) {
	flag := 0
	for elem := this.mostDB.Front(); elem != nil; elem = elem.Next() {
		cur := elem.Value.(WTopDBCost)
		if m.Cost < cur.Cost {
			this.mostDB.InsertBefore(m, elem)
			flag = 1
			break
		}
	}
	if flag == 0 {
		this.mostDB.PushBack(m)
	}
	if this.mostDB.Len() > 10 {
		this.mostDB.Remove(this.mostDB.Front())
	}
}

//到时间上报
func (this *monitor) commit() {
	this.report.Stamp = this.stamp
	this.report.RSS = GetMem()
	this.stamp = time.Now()   //更新下次循环到时
	this.report.SID = this.ID //记录为300秒前时间

	for elem := this.mostTime.Front(); elem != nil; elem = elem.Next() { //加入耗时top10
		this.report.TopRespond = append(this.report.TopRespond, elem.Value.(WRespond))
	}

	for elem := this.mostSize.Front(); elem != nil; elem = elem.Next() { //加入数据包top10
		topsize := elem.Value.(WTopSize)
		this.report.TopSize = append(this.report.TopSize, topsize)
	}
	for elem := this.mostDB.Front(); elem != nil; elem = elem.Next() { //加入数据包top10
		topdbcost := elem.Value.(WTopDBCost)
		this.report.TopDBCost = append(this.report.TopDBCost, topdbcost)
	}

	for _, v := range this.threadMap {
		var temp WThread
		temp.Name = v.name
		temp.MaxTime = int(v.max / time.Millisecond)
		temp.MinTime = int(v.min / time.Millisecond)
		temp.AllTime = int(v.total / time.Millisecond)
		temp.Ticks = v.tick
		this.report.Threads = append(this.report.Threads, temp)
	}
	LOG.INFO("monitor", "core monitor:", this.report.Stamp)
	sort.Sort(SortWThread(this.report.Threads))
	if this.CollConn != "" {
		WebReport(this.CollConn, WREPORT, this.report)
	} else {
		this.LogDB.Insert(DBSERVERLOG, COLCORERUNTIME, this.report)
		this.LogDB.Updatebyid(DBSERVERLOG, COLCORE, this.ID, bson.M{"rss": GetMem()})
	}
	//初始化
	this.mostTime = list.New()
	this.mostSize = list.New()
	this.mostDB = list.New()
	this.threadMap = make(map[string]*thread)
	this.report = WReport{}
	this.report.Peak = this.recentPeak //记录上次登录人员
}

//上报数据
func WebReport(url string, cmd int, data interface{}) []byte {
	postData := make([]byte, 0)
	if data != nil {
		switch data.(type) {
		case []byte:
			postData = data.([]byte)
		default:
			var err error
			postData, err = json.Marshal(data)
			if err != nil {
				LOG.ERROR("webreport error json marshal:", err.Error())
				return []byte("fail")
			}
		}
	}
	paddress := fmt.Sprintf("%s?cmd=%d", url, cmd)
	httpclient := &http.Client{Timeout: 10 * time.Second}
	rsp, er := httpclient.Post(paddress, "application/json;charset=utf-8", strings.NewReader(string(postData)))
	if er != nil {
		LOG.ERROR("webreport", "monitor", "error:", er.Error())
		return []byte("fail")
	}
	res, rer := ioutil.ReadAll(rsp.Body)
	if rer != nil {
		LOG.ERROR("webreport", "monitor", "error:", rer.Error())
		return []byte("fail")
	}
	return []byte(res)
}

func InitWLoginErr(code int) (this *WLoginErr) {
	this = new(WLoginErr)
	this.Code = code
	this.Nums = 1
	this.Detail = make([]LoginTStat, 0)
	switch this.Code {
	case LOGINERR_NOSERVERID: //1
		this.Desc = "core服务器不在线"
	case LOGINERR_TX: //2
		this.Desc = "腾讯认证错误"
	case LOGINERR_AGENTHTTP: //3
		this.Desc = "agent用户http请求出错"
	case LOGINERR_AGENTRECV: //4
		this.Desc = "agent用户http请求返回值错误"
	case LOGINERR_DBCONN: //5
		this.Desc = "数据库连接错误"
	case LOGINERR_NOTCP: //6
		this.Desc = "未发送TCP连接"
	case LOGINERR_CHECKERR: //7
		this.Desc = "Sig验证错误"
	case LOGINERR_COREERR: //8
		this.Desc = "core登录错误"
	}
	return
}

const (
	LOGINERR_SUCCESS    = 0 //登录成功
	LOGINERR_NOSERVERID = 1 //core服务器不在线
	LOGINERR_TX         = 2 //腾讯认证错误
	LOGINERR_AGENTHTTP  = 3 //agent http请求错误
	LOGINERR_AGENTRECV  = 4 //agent http验证结果错误
	LOGINERR_DBCONN     = 5 //数据库连接错误
	LOGINERR_NOTCP      = 6 //未发送TCP连接
	LOGINERR_CHECKERR   = 7 //验证错误
	LOGINERR_COREERR    = 8 //core未登录
)

//core合并
//oneSize 是否同一粒度
func (this *WReport) Merge(more []WReport, oneSize bool) {
	for _, one := range more {
		this.Out += one.Out
		this.In += one.In
		this.Responds += one.Responds
		if oneSize {
			this.Peak += one.Peak
		} else {
			if this.Peak < one.Peak {
				this.Peak = one.Peak
			}
		}
		this.TopRespond = append(this.TopRespond, one.TopRespond...)
		this.Threads = append(this.Threads, one.Threads...)
		this.TopSize = append(this.TopSize, one.TopSize...)
		this.TopDBCost = append(this.TopDBCost, one.TopDBCost...)
	}
	this.Threads = MergeThread(this.Threads)
	sort.Sort(SortWRespond(this.TopRespond))
	if len(this.TopRespond) > 10 {
		this.TopRespond = this.TopRespond[:10]
	}
	if len(this.TopSize) > 10 {
		sort.Sort(SortWTopSize(this.TopSize))
		this.TopSize = this.TopSize[:10]
	}
	if len(this.TopDBCost) > 10 {
		sort.Sort(SortWTopDBCost(this.TopDBCost))
		this.TopDBCost = this.TopDBCost[:10]
	}
}

//线程合并
func MergeThread(arr []WThread) []WThread {
	threadsMap := make(map[string]*WThread)
	for _, thread := range arr {
		if v, ok := threadsMap[thread.Name]; ok {
			v.AllTime += thread.AllTime
			v.Ticks += thread.Ticks
			if thread.MinTime < v.MinTime {
				v.MinTime = thread.MinTime
			}
			if thread.MaxTime > v.MaxTime {
				v.MaxTime = thread.MaxTime
			}
			v.MinTime += thread.MinTime
		} else {
			threadsMap[thread.Name] = &WThread{
				Name:    thread.Name,
				MaxTime: thread.MaxTime,
				MinTime: thread.MinTime,
				AllTime: thread.AllTime,
				Ticks:   thread.Ticks,
			}
		}
	}
	var newSlice = make([]WThread, 0)
	for _, v := range threadsMap {
		newSlice = append(newSlice, *v)
	}
	sort.Sort(SortWThread(newSlice))
	return newSlice
}

//gate合并
//oneSize是否同一粒度
func (this *WGateReport) Merge(more []WGateReport, oneSize bool) {
	for _, gate := range more {
		this.FlowIn += gate.FlowIn
		this.FlowOut += gate.FlowOut
		this.InNums += gate.InNums
		this.OutNums += gate.OutNums
		if oneSize {
			this.Peak += gate.Peak
		} else {
			if this.Peak < gate.Peak {
				this.Peak = gate.Peak
			}
		}
		this.LoginAll += gate.LoginAll
		this.LoginNums += gate.LoginNums
		this.Threads = append(this.Threads, gate.Threads...)
		this.TopRespond = append(this.TopRespond, gate.TopRespond...)
		this.TopSize = append(this.TopSize, gate.TopSize...)
		this.TopCmdNums = append(this.TopCmdNums, gate.TopCmdNums...)
		this.TopLogins = append(this.TopLogins, gate.TopLogins...)
		this.LoginErr = append(this.LoginErr, gate.LoginErr...)
	}
	//线程合并
	this.Threads = MergeThread(this.Threads)
	//耗时top10排序
	sort.Sort(SortWRespond(this.TopRespond))
	if len(this.TopRespond) > 10 {
		this.TopRespond = this.TopRespond[:10]
	}
	//包大小top10排序
	sort.Sort(SortWGateTopSize(this.TopSize))
	if len(this.TopSize) > 10 {
		this.TopSize = this.TopSize[:10]
	}
	//包相应命令数目top10排序,先合并
	CmdNumsMap := make(map[uint16]*WCmdNums)
	for _, obj := range this.TopCmdNums {
		if v, ok := CmdNumsMap[obj.Cmd]; ok {
			v.Nums += obj.Nums
		} else {
			CmdNumsMap[obj.Cmd] = &obj
		}
	}
	var CmdNumsSlice = make([]WCmdNums, 0)
	for _, v := range CmdNumsMap {
		CmdNumsSlice = append(CmdNumsSlice, *v)
	}
	sort.Sort(SortWCmdNums(CmdNumsSlice))
	if len(CmdNumsSlice) > 10 {
		this.TopCmdNums = CmdNumsSlice[:10]
	}
	//登录最慢top10
	sort.Sort(SortLoginTStat(this.TopLogins))
	if len(this.TopLogins) > 10 {
		this.TopLogins = this.TopLogins[:10]
	}
	//登录错误汇总
	this.LoginErr = MergeLoginerr(this.LoginErr)
}

//合并loginerr
func MergeLoginerr(arr []WLoginErr) []WLoginErr {
	loginErrMap := make(map[int]*WLoginErr)
	for _, obj := range arr {
		if v, ok := loginErrMap[obj.Code]; ok {
			v.Nums += obj.Nums
			v.Detail = append(v.Detail, obj.Detail...)
		} else {
			loginErrMap[obj.Code] = &obj
		}
	}
	var loginerrSlice = make([]WLoginErr, 0)
	for _, v := range loginErrMap {
		if len(v.Detail) > 10 {
			v.Detail = v.Detail[:10]
		}
		loginerrSlice = append(loginerrSlice, *v)
	}
	return loginerrSlice
}

//login合并
func (this *WLoginReport) Merge(more []WLoginReport, oneSize bool) {
	for _, one := range more {
		this.Responds += one.Responds
		this.Threads = append(this.Threads, one.Threads...)
		this.LoginErr = append(this.LoginErr, one.LoginErr...)
	}
	this.Threads = MergeThread(this.Threads)
	this.LoginErr = MergeLoginerr(this.LoginErr)
}

//排序

//包大小top10排序
type SortWTopDBCost []WTopDBCost

func (this SortWTopDBCost) Len() int           { return len(this) }
func (this SortWTopDBCost) Less(i, j int) bool { return this[i].Cost > this[j].Cost }
func (this SortWTopDBCost) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//包大小top10排序
type SortWTopSize []WTopSize

func (this SortWTopSize) Len() int           { return len(this) }
func (this SortWTopSize) Less(i, j int) bool { return this[i].Fsize > this[j].Fsize }
func (this SortWTopSize) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//线程名字字母排序
type SortWThread []WThread

func (this SortWThread) Len() int           { return len(this) }
func (this SortWThread) Less(i, j int) bool { return this[i].Name < this[j].Name }
func (this SortWThread) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//耗时top10
type SortWRespond []WRespond

func (this SortWRespond) Len() int           { return len(this) }
func (this SortWRespond) Less(i, j int) bool { return this[i].Cost > this[j].Cost }
func (this SortWRespond) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//包大小top10排序
type SortWGateTopSize []WGateTopSize

func (this SortWGateTopSize) Len() int           { return len(this) }
func (this SortWGateTopSize) Less(i, j int) bool { return this[i].Fsize > this[j].Fsize }
func (this SortWGateTopSize) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//包命令数目top10
type SortWCmdNums []WCmdNums

func (this SortWCmdNums) Len() int           { return len(this) }
func (this SortWCmdNums) Less(i, j int) bool { return this[i].Nums > this[j].Nums }
func (this SortWCmdNums) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//登录最慢top10
type SortLoginTStat []LoginTStat

func (this SortLoginTStat) Len() int           { return len(this) }
func (this SortLoginTStat) Less(i, j int) bool { return this[i].Cost > this[j].Cost }
func (this SortLoginTStat) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }

//获取内存大小
func GetMem() int {
	pid := os.Getpid()
	processInfo, err := process.NewProcess(int32(pid))
	if err == nil {
		mem, err1 := processInfo.MemoryInfo()
		if err1 == nil {
			return int(mem.RSS)
		} else {
			LOG.ERROR("monitor", "get memeory error:", err1.Error())
		}
	} else {
		LOG.ERROR("monitor", "get process error:", err.Error())
	}
	return 0
}
