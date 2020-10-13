package filelog

import (
	"errors"
	//	"fmt"
	"log"
	"os"
	"sync"
)

const (
	default_fileflag  int    = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	default_logprefix string = ""
	default_logflag   int    = log.LstdFlags
)

func InitFileLogger(filename string, params ...interface{}) (*log.Logger, error) {
	var fileflag int = default_fileflag
	if len(params) > 0 {
		fileflag = params[0].(int)
	}

	var logprefix string = default_logprefix
	if len(params) > 1 {
		logprefix = params[1].(string)
	}

	var logflag int = default_logflag
	if len(params) > 2 {
		logflag = params[2].(int)
	}

	if f, err := os.OpenFile(filename, fileflag, 0666); err == nil {
		return log.New(f, logprefix, logflag), nil
	} else {
		return nil, err
	}
}

const (
	FILELOG_DEBUG  = 1 << iota //指定日志输出debug内容
	FILELOG_STDOUT             //指定日志输出标准输出内容
	FILELOG_NOFILE             //指定日志不产生任何自定义文件
)

//type FileLog struct {
//	err  *log.Logger
//	info *log.Logger
//	std  *log.Logger

//	name string
//	flag int
//	l    *sync.Mutex //打开err.log时使用
//}

var (
	//	//初始化日志
	//	INITLOG *NewFileLog = nil
	//默认日志
	LOG *NewFileLog = &NewFileLog{"./", nil, nil, "default", FILELOG_NOFILE | FILELOG_STDOUT, &sync.Mutex{}, sync.RWMutex{}, _getnextrotatetime()}
	//默认日志是否被重置过
	boinit bool
)

//func init() {
//	INITLOG = CreateFileLog("init", FILELOG_STDOUT)
//	if err := INITLOG.Init("init", FILELOG_STDOUT); err != nil {
//		panic(err.Error())
//	}
//}

/**默认日志*/
func INFO(v ...interface{}) {
	LOG.INFO(v...)
}

func ERROR(v ...interface{}) {
	LOG.ERROR(v...)
}

func DEBUG(v ...interface{}) {
	LOG.DEBUG(v...)
}

func FATAL(v ...interface{}) {
	LOG.FATAL(v...)
}

//初始化默认日志 应该只被初始化一次  目前已被初始化的文件句柄未做回收处理 所以这里的多次初始化会被视为无效，需要支持多次变更默认Logger的时候可以加上相应释放旧文件句柄的语句
func Init(name string, flag int, logpath string, filelocker sync.Locker) error {
	if boinit {
		return errors.New("filelog has been init.")
	}
	boinit = true
	if l, err := CreateFileLog(name, flag, logpath, filelocker); err != nil {
		return err
	} else {
		LOG = l
		return nil
	}
}

//func (f *FileLog) Init(name string, flag int) error {
//	f.name = name
//	f.flag = flag
//	f.l = new(sync.Mutex)
//	if l, err := InitFileLogger("./" + name + ".log"); err != nil {
//		return err
//	} else {
//		f.info = l
//	}

//	f.std = log.New(os.Stdout, "", default_logflag)

//	return nil
//}

//func (f *FileLog) INFO(v ...interface{}) {
//	outs := ""
//	if f.flag&FILELOG_NOFILE == 0 {
//		if len(outs) == 0 {
//			outs = "INFO " + fmt.Sprintln(v...)
//		}
//		f.info.Output(1, outs)
//	}

//	if f.flag&(FILELOG_STDOUT|FILELOG_DEBUG) != 0 {
//		if len(outs) == 0 {
//			outs = "INFO " + fmt.Sprintln(v...)
//		}
//		f.std.Output(1, outs)
//	}
//}

//func (f *FileLog) ERROR(v ...interface{}) {
//	outs := ""
//	if f.flag&FILELOG_NOFILE == 0 {
//		if len(outs) == 0 {
//			outs = "ERROR " + fmt.Sprintln(v...)
//		}
//		f.info.Output(1, outs)
//		if f.err != nil {
//			f.err.Output(1, outs)
//		} else {
//			f.l.Lock()
//			defer f.l.Unlock()
//			if l, err := InitFileLogger("./err.log"); err != nil {
//				log.Println("open err.log fail , errinfo : ", err.Error())
//			} else {
//				f.err = l
//				f.err.Output(1, outs)
//			}
//		}
//	}
//	if f.flag&(FILELOG_STDOUT|FILELOG_DEBUG) != 0 {
//		if len(outs) == 0 {
//			outs = "ERROR " + fmt.Sprintln(v...)
//		}
//		f.std.Output(1, outs)
//	}

//	if f.flag&FILELOG_ERRFATAL != 0 {
//		if len(outs) == 0 {
//			outs = "ERROR " + fmt.Sprintln(v...)
//		}
//		panic(outs)
//	}
//}

//func (f *FileLog) DEBUG(v ...interface{}) {
//	if f.flag&FILELOG_DEBUG != 0 {
//		outs := "DEBUG " + fmt.Sprintln(v...)
//		f.std.Output(1, outs)
//		if f.flag&FILELOG_NOFILE == 0 {
//			f.info.Output(1, outs)
//		}
//	}
//}
