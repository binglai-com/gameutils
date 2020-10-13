package trans

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

//事务结构
type Trans struct {
	Tid   string        `bson:"_id"`   //事务唯一编号
	Type  uint16        `bson:"typ"`   //事务类型 用户自定义
	Param []interface{} `bson:"param"` //事务参数 用户自定义
	Ct    time.Time     `bson:"ct"`    //事务创建时间
	St    time.Time     `bson:"st"`    //事务受理时间
	Ot    time.Time     `bson:"ot"`    //事务完结时间
	Ret   int           `bson:"ret"`   //事务处理返回码 初始为-1表示未完结
	Res   string        `bson:"res"`   //事务处理返回描述 初始为空字符串
}

//新建一条事务记录
func NewTrans(TranTyp uint16, params ...interface{}) *Trans {
	return &Trans{
		bson.NewObjectId().Hex(),
		TranTyp,
		params[:],
		time.Now(),
		time.Time{},
		time.Time{},
		-1,
		""}
}

//事务开始受理
func (t *Trans) Start(n time.Time) {
	t.St = n
}

//事务完结
func (t *Trans) Over(ret int, res string) {
	t.Ot, t.Ret, t.Res = time.Now(), ret, res
}
