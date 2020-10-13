package gmhelper

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"sync"
)

const (
	GMERR_EXECFAIL = "cmdexefail"
)

var (
	CmdNotFound  = fmt.Errorf("cmd not found")
	InvalidParam = fmt.Errorf("invalidparam")
)

//Gm命令描述
type GmCmdDesc struct {
	Cmd  string `json:"cmd"`
	Desc string `json:"desc"`
}

//Gm命令处理程序
type GmHandler struct {
	Desc        string                                //命令描述和参数描述
	Implementer func(params GmParams) (error, string) //Gm命令实现程序
}

var DefaultGmHelper *GmHelper

func init() {
	DefaultGmHelper = NewGmHelper()
}

func HandleFunc(cmdname string, cmddesc string, f func(params GmParams) (error, string)) {
	DefaultGmHelper.HandleFunc(cmdname, cmddesc, f)
}

func GetAllCmdDesc() (ret []GmCmdDesc) {
	return DefaultGmHelper.GetAllCmdDesc()
}

func ExecCmd(cmd string, params GmParams) (error, string) {
	return DefaultGmHelper.ExecCmd(cmd, params)
}

func GetCmdDesc(cmd string) string {
	return DefaultGmHelper.GetCmdDesc(cmd)
}

//Gm命令工具
type GmHelper struct {
	l       *sync.Mutex //保护对以下内容的访问
	hanlder map[string]GmHandler
}

//创建一个Gm命令工具
func NewGmHelper() *GmHelper {
	h := new(GmHelper)
	h.Init()
	return h
}

//命令工具初始化
func (g *GmHelper) Init() {
	g.l = new(sync.Mutex)
	g.hanlder = make(map[string]GmHandler)
}

//增加命令程序
func (g *GmHelper) HandleFunc(cmdname string, cmddesc string, f func(params GmParams) (error, string)) {
	g.l.Lock()
	defer g.l.Unlock()
	g.hanlder[cmdname] = GmHandler{cmddesc, f}
}

//获取命令描述
func (g *GmHelper) GetAllCmdDesc() (ret []GmCmdDesc) {
	g.l.Lock()

	for cmdname, v := range g.hanlder {
		var t GmCmdDesc
		t.Cmd = cmdname
		t.Desc = v.Desc
		ret = append(ret, t)
	}

	g.l.Unlock()

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Cmd < ret[j].Cmd
	})

	return
}

type GmParams map[string]string //Gm函数参数

func (this GmParams) GetInt(name string, defaulti int) int {
	sn := this[name]
	if sn == "" {
		return defaulti
	} else {
		if tmp, err := strconv.Atoi(sn); err != nil {
			return defaulti
		} else {
			return tmp
		}
	}
}

func (this GmParams) GetString(name string, defaults string) string {
	if tmp, ok := this[name]; ok {
		return tmp
	} else {
		return defaults
	}
}

func (this GmParams) Have(name string) bool {
	_, ok := this[name]
	return ok
}

//检查遗漏的参数列表
func (this GmParams) GetOmit(keys ...string) string {
	for _, key := range keys {
		if !this.Have(key) {
			return key
		}
	}
	return ""
}

//根据URL解析执行参数
func GmParamsFromUrl(URL string) GmParams {
	ret := make(GmParams)
	tmpurl, err := url.Parse(URL)
	if err != nil {
		return ret
	}

	vals, parseerr := url.ParseQuery(tmpurl.RawQuery)
	if parseerr != nil {
		return ret
	}

	for k, _ := range vals {
		ret[k] = vals.Get(k)
	}
	return ret
}

//执行命令
func (g *GmHelper) ExecCmd(cmd string, params GmParams) (err error, res string) {
	var handler GmHandler
	g.l.Lock()
	if h, ok := g.hanlder[cmd]; !ok {
		g.l.Unlock()
		err = CmdNotFound
		return
	} else {
		g.l.Unlock()
		handler = h
	}

	return handler.Implementer(params)
}

//获得命令描述
func (g *GmHelper) GetCmdDesc(cmd string) string {
	g.l.Lock()
	defer g.l.Unlock()
	if h, ok := g.hanlder[cmd]; ok {
		return h.Desc
	}
	return ""
}

//格式化命令列表
func FormatCmdList(cmdlist []GmCmdDesc) (ret string) {
	for _, v := range cmdlist {
		ret += v.Cmd + "\t:\t" + v.Desc + "\n"
	}
	return
}
