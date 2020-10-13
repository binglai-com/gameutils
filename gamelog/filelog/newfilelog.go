package filelog

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

//新文件日志
type NewFileLog struct {
	logpath string   //日志目录
	info    *os.File //info、debug、err 文本日志
	err     *os.File //err 错误日志
	name    string   //日志名称
	flag    int      //标记

	filelock   sync.Locker  //文件锁
	l          sync.RWMutex //内部信息读写锁
	nextrotate time.Time    //下次日志交替的时间
}

//获取下次日志交替的时间
func _getnextrotatetime() time.Time {
	var now = time.Now()
	var now3clock = time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if now.After(now3clock) {
		return now3clock.AddDate(0, 0, 1)
	} else {
		return now3clock
	}
}

//格式化时间缓存
type formatCacheType struct {
	LastUpdateSeconds int64
	TimeFormat        string
}

var formatCache = &formatCacheType{}

//创建文件日志
func CreateFileLog(name string, flag int, logpath string, filelocker sync.Locker) (*NewFileLog, error) {
	//检查日志路径
	if filelocker == nil {
		filelocker = &sync.Mutex{} //默认锁类型
	}

	if info, err := os.Stat(logpath); err != nil { //日志路径不存在 创建路径
		if perr := os.MkdirAll(logpath, 0666); perr != nil { //创建日志路径失败
			return nil, perr
		}
	} else {
		if !info.IsDir() { //目标不是路径
			return nil, fmt.Errorf("the logpath must be a valid dir.")
		}
	}

	if fd, err := os.OpenFile(path.Join(logpath, fmt.Sprintf("%s.log", name)), default_fileflag, 0666); err != nil {
		return nil, err
	} else {
		return &NewFileLog{
			logpath,
			fd,
			nil,
			name,
			flag,
			filelocker,
			sync.RWMutex{},
			_getnextrotatetime()}, nil
	}
}

const (
	_prifix_debug = "DEBUG "
	_prifix_error = "ERROR "
	_prifix_info  = "INFO "
	_prifix_fatal = "FATAL "
)

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int, bufstart int) {
	// Assemble decimal in reverse order.
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		(*buf)[bufstart+wid] = byte('0' + i - q*10)
		i = q
	}
	// i < 10
	(*buf)[bufstart] = byte('0' + i)
}

//日志更新
func (l *NewFileLog) _update() {
	var now = time.Now()
	var secs = now.UnixNano() / 1e9
	if formatCache.LastUpdateSeconds != secs {
		var buf = make([]byte, 20)
		year, month, day := now.Date()
		hour, min, sec := now.Clock()
		itoa(&buf, year, 4, 0)
		buf[4] = '/'
		itoa(&buf, int(month), 2, 5)
		buf[7] = '/'
		itoa(&buf, day, 2, 8)
		buf[10] = ' '
		itoa(&buf, hour, 2, 11)
		buf[13] = ':'
		itoa(&buf, min, 2, 14)
		buf[16] = ':'
		itoa(&buf, sec, 2, 17)
		buf[19] = ' '

		var updated = &formatCacheType{
			LastUpdateSeconds: secs,
			TimeFormat:        string(buf)}

		formatCache = updated
	}

	//检查是否到达下次日志交替的时间
	l.l.RLock()
	var nextrotate = l.nextrotate
	l.l.RUnlock()
	if now.After(nextrotate) {
		l.l.Lock()
		defer l.l.Unlock()
		l.nextrotate = _getnextrotatetime()                                             //更新下次滚动日期
		var newname = fmt.Sprintf("%s_%s", l.name, nextrotate.Format("20060102150405")) //本次滚动的日志名称
		l.info.Close()                                                                  //关闭旧文件 准备滚动日志

		//开始滚动日志文件
		l.filelock.Lock() //文件锁
		defer l.filelock.Unlock()
		//判断是否其他进程已经滚动完毕
		var _, err1 = os.Stat(path.Join(l.logpath, fmt.Sprintf("%s.log", newname)))
		var _, err2 = os.Stat(path.Join(l.logpath, fmt.Sprintf("%s.zip", newname)))
		if err1 != nil && err2 != nil { //两个都不存在 说明rotate操作尚未开始，因此这里开始执行rotate操作
			fmt.Println("开始执行rotate操作")
			for { //等待其他进程释放该文件句柄
				if err := os.Rename(path.Join(l.logpath, fmt.Sprintf("%s.log", l.name)), path.Join(l.logpath, fmt.Sprintf("%s.log", newname))); err == nil {
					fmt.Println("rename file success!!")
					break
				} else {
					time.Sleep(time.Millisecond * 10)
				}

			}

			go func() { //压缩文件
				if fi, er1 := os.OpenFile(path.Join(l.logpath, fmt.Sprintf("%s.log", newname)), os.O_RDONLY, 0666); er1 == nil {
					defer fi.Close()
					if fo, er2 := os.OpenFile(path.Join(l.logpath, fmt.Sprintf("%s.zip", newname)), os.O_CREATE|os.O_RDWR|os.O_EXCL, 0666); er2 == nil {
						defer fo.Close()
						var zipwriter = zip.NewWriter(fo)
						if wr, er3 := zipwriter.Create(newname); er3 == nil {
							if _, er4 := io.Copy(wr, fi); er4 == nil {
								if er5 := zipwriter.Close(); er5 == nil {
									fi.Close()
									if er6 := os.Remove(path.Join(l.logpath, fmt.Sprintf("%s.log", newname))); er6 != nil {
										fmt.Println("zip log file rm old file fail : ", er6.Error())
									}
									//压缩成功
									fmt.Println("zip log file complete!")
								} else {
									fmt.Println("zip log file zipwriter close fail : ", er5.Error())
								}
							} else {
								fmt.Println("zip log file io.copy fail : ", er4.Error())
							}
						} else {
							fmt.Println("zip log file create zipwriter fail : ", er3.Error())
						}
					} else {
						fmt.Println("zip log file create zipfile fail : ", er2.Error())
					}
				} else {
					fmt.Println("zip log file open oldfile fail : ", er1.Error())
				}
			}()
		} else { // rotate产生的.zip或.log任意一个存在 说明 rotate工作已经完成 这里不再执行rotate操作
			fmt.Println("rotate操作已执行完毕  更新当前info fd")
		}

		if fd, err := os.OpenFile(path.Join(l.logpath, fmt.Sprintf("%s.log", l.name)), default_fileflag, 0666); err != nil {
			fmt.Println("open log file fail : ", err.Error())
		} else {
			l.info = fd
		}
	}
}

//格式刷输出文件
func (l *NewFileLog) FormatLog(bosource bool, prefix string, v ...interface{}) []byte {
	var cache = *formatCache
	var out = bytes.NewBuffer(make([]byte, 0, 64))
	out.WriteString(cache.TimeFormat)
	out.WriteString(prefix)

	if bosource {
		pc, _, lineno, ok := runtime.Caller(3)
		src := ""
		if ok {
			src = fmt.Sprintf("%s:%d ", runtime.FuncForPC(pc).Name(), lineno)
		}
		out.WriteString(src)
	}

	out.WriteString(fmt.Sprintln(v...))

	return out.Bytes()
}

func (l *NewFileLog) DEBUG(v ...interface{}) {
	l._update()
	if l.flag&FILELOG_DEBUG != 0 {
		//判断是否输出文件
		var outs = l.FormatLog(false, _prifix_debug, v...)
		os.Stdout.Write(outs)
		if l.flag&FILELOG_NOFILE == 0 {
			l.info.Write(outs)
		}
	}
}

func (l *NewFileLog) FATAL(v ...interface{}) {
	l._update()
	var outs = []byte{}
	if l.flag&FILELOG_NOFILE == 0 {
		outs = l.FormatLog(true, _prifix_fatal, v...)
		l.info.Write(outs)
		if l.err == nil {
			//创建错误日志文件
			if fd, err := os.OpenFile(path.Join(l.logpath, fmt.Sprintf("%s_err.log", l.name)), default_fileflag, 0666); err != nil {
				log.Println("open err.log fail , errinfo : ", err.Error())
			} else {
				l.err = fd
			}
		}
		l.err.Write(outs)
	}

	if l.flag&(FILELOG_STDOUT|FILELOG_DEBUG) != 0 {
		if len(outs) == 0 {
			outs = l.FormatLog(true, _prifix_fatal, v...)
		}
		os.Stdout.Write(outs)
	}

	if len(outs) == 0 {
		outs = l.FormatLog(true, _prifix_fatal, v...)
	}
	panic(string(outs))
}

func (l *NewFileLog) ERROR(v ...interface{}) {
	l._update()
	var outs = []byte{}
	if l.flag&FILELOG_NOFILE == 0 {
		outs = l.FormatLog(true, _prifix_error, v...)
		l.info.Write(outs)
		if l.err == nil {
			//创建错误日志文件
			if fd, err := os.OpenFile(path.Join(l.logpath, fmt.Sprintf("%s_err.log", l.name)), default_fileflag, 0666); err != nil {
				log.Println("open err.log fail , errinfo : ", err.Error())
			} else {
				l.err = fd
			}
		}
		l.err.Write(outs)
	}

	if l.flag&(FILELOG_STDOUT|FILELOG_DEBUG) != 0 {
		if len(outs) == 0 {
			outs = l.FormatLog(true, _prifix_error, v...)
		}
		os.Stdout.Write(outs)
	}
}
func (l *NewFileLog) INFO(v ...interface{}) {
	l._update()
	var outs = []byte{}
	if l.flag&FILELOG_NOFILE == 0 {
		if len(outs) == 0 {
			outs = l.FormatLog(false, _prifix_info, v...)
		}
		l.info.Write(outs)
	}

	if l.flag&(FILELOG_STDOUT|FILELOG_DEBUG) != 0 {
		if len(outs) == 0 {
			outs = l.FormatLog(false, _prifix_info, v...)
		}
		os.Stdout.Write(outs)
	}
}
