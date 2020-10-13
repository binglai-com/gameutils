package gamemonitor

import (
	"time"
)

const (
	WCORESTART      = 1  //游戏开始 WCoreStart
	WCOREEND        = 2  //游戏结束
	WREPORT         = 3  //上报监控内容
	WGATESTART      = 4  //网关启动 WGateStart
	WGATEREPORT     = 5  //网关汇报
	WLOGINSTART     = 6  //webservice启动 WLoginStart
	WLOGINREPORT    = 7  //webservice汇报
	WGATEEND        = 8  //网关停止
	WLOGINEND       = 9  //webservice结束
	WGETCORESETTING = 10 //获取core配置
	WOPENSER        = 11 //收到服务器列表有更新通知
	WACTIVITY       = 12 //收到活动信息
	WUPDATEIP       = 13 //更新IP
	WUSERTRAFFIC    = 14 //用户流量汇报
	WUPDCORE        = 15 //维护core
	WUPDGATE        = 16 //维护gate
	WUPDLOGIN       = 17 //维护login
	WGETUPDBYID     = 18 //获取维护的信息
	WREBOOT         = 19 //重启
	WDANGER         = 20 //危险信息
)

type WDanger struct {
	Type   int    //分类
	Desc   string //描述
	Detail string //明细
}

const (
	DBLOADING       = "dbloading"
	DBSERVERLOG     = "web_sermgr"
	COLGATELOADING  = "gateloading"
	COLLOGINLOADING = "loginloading"
	COLCORE         = "core"
	COLGATE         = "gate"
	COLLOGIN        = "login"
	COLCORERUNTIME  = "core_runtime"
	COLGATERUNTIME  = "gate_runtime"
	COLLOGINRUNTIME = "login_runtime"
)

//core
type WCoreStart struct {
	ID            string `json:"ID"bson:"_id"`
	Gateid        string `json:"gateid"bson:"gateid"`
	Serverid      []int  `json:"Serverid"bson:"serverid"`
	TcpConn       string `json:"TcpConn"bson:"tcpconn"`
	CollConn      string `json:"CollConn"bson:"collconn"`
	DBPlayerConn  string `json:"DBPlayerConn"bson:"coreconn"`
	DBGLogConn    string `json:"DBGLogConn"bson:"logconn"`
	GameBase      string `json:"GameBase"bson:"gamebase"`
	Coreport      int    `json:"Coreport"bson:"coreport"`           //ip地址
	TcpPort       int    `json:"tcpport"bson:"tcpport"`             //TCP端口
	BoxPort       int    `json:"boxport"bson:"boxport"`             //沙箱端口
	ServerProp    int    `json:"ServerProp"bson:"serverprop"`       // 1正式服/0测试服
	ConfigVersion string `json:"configversion"bson:"configversion"` //配置表版本
	//以上是core Config配置
	Ip       string    `json:"ip"bson:"ip"`             //代理访问IP
	Version  string    `json:"version"bson:"version"`   //版本
	InsideIp string    `json:"insideip"bson:"insideip"` //内部访问IP
	State    int       `json:"state"bson:"state"`       //0正常运行
	Stamp    time.Time `json:"-"bson:"-"`               //上次收到汇报时间
	SigConn  string    `json:"sigconn"bson:"sigconn"`   //sig认证地址
	Opentime time.Time `json:"opentime"bson:"opentime"` //开服时间
}

//gate
type WGateStart struct {
	ID       string `json:"ID"bson:"_id"`
	HttpPort int    `json:"HttpPort"bson:"httpport"`
	CollConn string `json:"CollConn"bson:"collconn"`
	TcpPort  int    `json:"TcpPort"bson:"tcpport"`
	BoxPort  int    `json:"BoxPort"bson:"boxport"`
	//以上是gate config配置
	Ip       string    `json:"ip"bson:"ip"`             //代理访问IP
	Version  string    `json:"version"bson:"version"`   //版本
	InsideIp string    `json:"insideip"bson:"insideip"` //内部访问IP
	State    int       `json:"state"bson:"state"`       //0正常运行
	Stamp    time.Time `json:"-"bson:"-"`               //上次收到汇报时间
}

//login
type WLoginStart struct {
	ID          string `json:"id"bson:"_id"`                  //webservice编号
	CollConn    string `json:"collconn"bson:"collconn"`       //web连接
	HttpPort    int    `json:"httpport"bson:"httpport"`       //http内部接口
	DBLoginConn string `json:"dbloginconn"bson:"dbloginconn"` //webservice数据库
	DBWebConn   string `json:"dbwebconn"bson:"dbwebconn"`     //平台库
	WebAgent    string `json:"webagent"bson:"webagent"`
	GameBase    string `json:"gamebase"bson:"gamebase"`
	//以上是login config配置
	Version  string    `json:"version"bson:"version"`   //版本
	State    int       `json:"state"bson:"state"`       //0正常运行
	Stamp    time.Time `json:"-"bson:"-"`               //上次收到汇报时间
	InsideIp string    `json:"insideip"bson:"insideip"` //内部访问IP
	ConnType int       `json:"conntype"bson:"conntype"` //采集器连接类型
}

type WLoginErr struct {
	Code   int          `json:"code"bson:"code"`
	Nums   int          `json:"nums"bson:"nums"`
	Desc   string       `json:"desc"bson:"desc"`
	Detail []LoginTStat `json:"detail"bson:"detail"`
}

type WCmdNums struct {
	Cmd  uint16 `json:"cmd"bson:"cmd"`
	Nums int    `json:"nums"bson:"nums"`
	Flag int    `json:"flag"bson:"flag"`
}

type WGateReport struct {
	SID        string         `json:"sid"bson:"sid"`               //网关编号
	Stamp      time.Time      `json:"stamp"bson:"stamp"`           //时间点
	FlowIn     int            `json:"flowIn"bson:"flowin"`         //流入
	FlowOut    int            `json:"flowOut"bson:"flowout"`       //流出
	InNums     int            `json:"innums"bson:"innums"`         //处理指令入个数
	OutNums    int            `json:"outnums"bson:"outnums"`       //处理指令出个数
	Peak       int            `json:"peak"bson:"peak"`             //在线客户端峰值
	Threads    []WThread      `json:"threads"bson:"threads"`       //线程相关
	TopRespond []WRespond     `json:"toprespond"bson:"toprespond"` //耗时TOP10
	TopSize    []WGateTopSize `json:"topsize"bson:"topsize"`       //包大小TOP10
	TopCmdNums []WCmdNums     `json:"topcmdnums"bson:"topcmdnums"` //命令个数
	TopLogins  []LoginTStat   `json:"toplogins"bson:"toplogins"`   //登录最慢TOP10
	LoginAll   int            `json:"loginall"bson:"loginall"`     //登录总耗时
	LoginNums  int            `json:"loginnums"bson:"loginnums"`   //登录总次数
	LoginErr   []WLoginErr    `json:"loginerr"bson:"loginerr"`     //登录错误统计
	RSS        int            `json:"rss"bson:"-"`                 //消耗内存
}

//数据包
type WGateTopSize struct {
	Cmd   uint16 `json:"cmd"bson:"cmd"`
	Fsize int    `json:"fsize"bson:"fsize"` //大小
	Flag  int    `json:"flag"bson:"flag"`   //0入1出
}

type WReport struct {
	SID        string       `json:"sid"bson:"sid"`               //服务器编号
	Stamp      time.Time    `json:"stamp"bson:"stamp"`           //提交时间
	Out        int          `json:"out"bson:"out"`               //字节数(B)
	In         int          `json:"in"bson:"in"`                 //字节数(B)
	Responds   int          `json:"responds"bson:"responds"`     //处理指令个数
	TopRespond []WRespond   `json:"toprespond"bson:"toprespond"` //耗时最多的10个协议
	Peak       int          `json:"peak"bson:"peak"`             //在线人数峰值
	Threads    []WThread    `json:"threads"bson:"threads"`       //线程相关
	TopSize    []WTopSize   `json:"topsize"bson:"topsize"`       //最大的10个数据包
	TopDBCost  []WTopDBCost `json:"topdbcost"bson:"topdbcost"`   //数据库耗时最大10个
	RSS        int          `json:"rss"bson:"-"`                 //消耗内存
}

//数据包
type WTopSize struct {
	Code  int `json:"code"bson:"code"`   //值
	Fsize int `json:"fsize"bson:"fsize"` //大小
	Flag  int `json:"flag"bson:"flag"`   //0入;1出
}

//db花费
type WTopDBCost struct {
	Cost   int    `json:"cost"bson:"cost"`
	Optype uint16 `json:"optype"bson:"optype"`
	Dbtype int    `json:"dbtype"bson:"dbtype"`
	Record string `json:"record"bson:"record"`
}

type WThread struct {
	Name    string `json:"name"bson:"name"`       //线程名
	MaxTime int    `json:"maxtime"bson:"maxtime"` //最大耗时
	MinTime int    `json:"mintime"bson:"mintime"` //最小耗时
	AllTime int    `json:"alltime"bson:"alltime"` //总耗时
	Ticks   int    `json:"ticks"bson:"ticks"`     //总次数
}

type WRespond struct {
	Cmd  uint16 `json:"cmd"bson:"cmd"`   //命令号
	Flag int    `json:"flag"bson:"flag"` //出入标记
	Data string `json:"data"bson:"data"` //数据
	Cost int    `json:"cost"bson:"cost"` //花费
}

type WLoginReport struct {
	SID      string      `json:"sid"bson:"sid"`           //webservice编号
	Stamp    time.Time   `json:"stamp"bson:"stamp"`       //时间点
	Responds int         `json:"responds"bson:"responds"` //处理登录个数
	Threads  []WThread   `json:"threads"bson:"threads"`   //线程相关
	LoginErr []WLoginErr `json:"loginerr"bson:"loginerr"` //登录错误统计
	RSS      int         `json:"rss"bson:"-"`             //消耗内存
}

//登录统计
type LoginTStat struct {
	Sessionid uint32        `json:"sess"bson:"sess"`         //会话号
	Openid    string        `json:"openid"bson:"openid"`     //openid
	Account   string        `json:"account"bson:"account"`   //账号
	Cost      time.Duration `json:"cost"bson:"cost"`         //总花费时间
	PF        string        `json:"pf"bson:"pf"`             //平台
	Nick      string        `json:"nick"bson:"nick"`         //昵称
	Vip       int           `json:"vip"bson:"vip"`           //vip等级
	Serverid  int           `json:"serverid"bson:"serverid"` //服务器号
	StartTime time.Time     `json:"start"bson:"start"`       //开始登录时间
	ErrCode   int           `json:"err"bson:"err"`           //错误码
	Sig       string        `json:"sig"bson:"sig"`           //验证码
	CheckT    time.Time     `json:"checkt"bson:"checkt"`     //验证时间
	GateT     time.Time     `json:"gatet"bson:"gatet"`       //gate收到时间
}

//gate加载日志
type GateLoadingLog struct {
	Ret          int           `json:"ret"bson:"ret"`           //登录结果
	IsNew        int           `json:"isnew"bson:"isnew"`       //是否新用户
	Sessionid    uint32        `json:"sess"bson:"sess"`         //会话号
	Openid       string        `json:"openid"bson:"openid"`     //openid
	Account      string        `json:"account"bson:"account"`   //账号
	Serverid     int           `json:"serverid"bson:"serverid"` //服务器号
	Gateid       string        `json:"gateid"bson:"gateid"`     //ID号
	Cost         time.Duration `json:"cost"bson:"cost"`         //登录花费时间
	PF           string        `json:"pf"bson:"pf"`             //平台
	Vip          int           `json:"vip"bson:"vip"`           //vip等级
	YearVIP      int           `json:"yearvip"bson:"yearvip"`   //年费VIP
	SuperVIP     int           `json:"supervip"bson:"supervip"` //豪华VIP
	IsVIP        int           `json:"isvip"bson:"isvip"`       //是否VIP
	CTime        time.Time     `json:"ctime"bson:"ctime"`       //开始登录时间
	Sig          string        `json:"sig"bson:"sig"`           //验证码
	GateT        time.Time     `json:"gatet"bson:"gatet"`       //gate收到时间
	CheckT       time.Time     `json:"checkt"bson:"checkt"`     //验证时间
	NewUserT     time.Time     `bson:"newusert"`                //通知新用户时间
	CreateT      time.Time     `bson:"createt"`                 //收到创建角色时间
	SLoginT      time.Time     `bson:"slogint"`                 //收到登录成功时间
	OfflineT     time.Time     `bson:"offlinet"`                //下线时间
	InFlow       int           `bson:"inflow"`                  //入流量
	OutFlow      int           `bson:"outflow"`                 //出流量
	Pid          string        `bson:"pid"`                     //玩家ID
	LoginID      string        `bson:"loginid"`                 //登录ID
	UserIP       string        `bson:"userip"`                  //用户IP
	LoaddingType int           `bson:"loadingtype"`             //加载方式 0新加载 1版本更新 2重新加载
}

type RegistLog struct {
	Openid     string    `json:"openid"bson:"openid"`         //openid
	Account    string    `json:"account"bson:"account"`       //账号
	Serverid   int       `json:"serverid"bson:"serverid"`     //服务器号
	Pid        string    `json:"pid"bson:"pid"`               //玩家ID
	PF         string    `json:"pf"bson:"pf"`                 //平台
	Vip        int       `json:"vip"bson:"vip"`               //vip等级
	YearVIP    int       `json:"yearvip"bson:"yearvip"`       //年费VIP
	SuperVIP   int       `json:"supervip"bson:"supervip"`     //豪华VIP
	IsVIP      int       `json:"isvip"bson:"isvip"`           //是否VIP
	Time       time.Time `json:"time"bson:"time"`             //登录时间
	ClientCost int64     `json:"clientcost"bson:"clientcost"` //客户端耗费时间
	Cost       int64     `json:"cost"bson:"cost"`             //耗费总时间
	UserIP     string    `bson:"userip"`                      //用户IP
}

//gate base
type GateBase struct {
	Id    string
	RConn string
	Peal  int
}

//core base
type CoreBase struct {
	Serverid []int
	RConn    string
	HttpConn string
	GateId   string
}

type AllSerBase struct {
	G []*GateBase
	C []*CoreBase
}

//采集器返回消息
const (
	COLL_EXIST     = 1 //已经存在
	COLL_AC        = 2 //添加core消息
	COLL_SERLIST   = 3 //开服列表信息
	COLL_SIG       = 4 //sig认证消息
	COLL_UPDCONFIG = 5 //更新配置
	COLL_OPENTIME  = 6 //开服时间
	COLL_REBOOT    = 7 //重启
)
