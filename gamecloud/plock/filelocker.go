package plock

import (
	"fmt"
	"os"
	"path"
	"time"
)

type FileLocker struct {
	path string
}

func NewFileLocker(lockpath string) (*FileLocker, error) {
	if stat, err := os.Stat(lockpath); err != nil {
		if !os.IsExist(err) { //目标路径不存在  创建目标路径
			if direrr := os.MkdirAll(lockpath, 0666); direrr != nil {
				return nil, fmt.Errorf("init filelocker lockpath fail : %s", direrr.Error())
			}
		} else { //其他错误
			return nil, fmt.Errorf("unexpected stat err : %s", err.Error())
		}
	} else { //路径存在
		if !stat.IsDir() { //目标参数不是路径
			return nil, fmt.Errorf("the input must be a valid dir path.")
		}
	}
	return &FileLocker{path.Join(lockpath, "plock")}, nil
}

func (l *FileLocker) Lock() {
	for {
		if f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL, 0666); err != nil {
			time.Sleep(time.Millisecond * 10)
		} else { //获得锁
			f.Close() //文件打开后要关闭
			break
		}
	}
}

func (l *FileLocker) Unlock() {
	os.Remove(l.path)
}
