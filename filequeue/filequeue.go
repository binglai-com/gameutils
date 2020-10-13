package filequeue

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

//文件队列
type FileQueque struct {
	name       string
	curindex   int64
	l_data     sync.RWMutex
	f_w_data   *os.File //数据文件写入
	f_r_data   *os.File //数据文件读取
	f_data_idx *os.File //数据索引文件

	rlock sync.Mutex //读锁
}

//初始化queue 传入数据路径
func InitQueue(name string) (*FileQueque, error) {
	var que = new(FileQueque)
	que.name = name
	if f_w_data, err := os.OpenFile("./"+name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		return nil, err
	} else {
		que.f_w_data = f_w_data
	}

	if f_r_data, err := os.OpenFile("./"+name, os.O_RDONLY, 0666); err != nil {
		return nil, err
	} else {
		que.f_r_data = f_r_data
	}

	if f_data_idx, err := os.OpenFile("./"+name+".idx", os.O_CREATE|os.O_RDWR, 0666); err != nil {
		return nil, err
	} else {
		que.f_data_idx = f_data_idx
	}

	if sindex, err := ioutil.ReadAll(que.f_data_idx); err != nil {
		return nil, err
	} else {
		if len(sindex) != 0 {
			if i, err := strconv.ParseInt(string(sindex), 10, 64); err != nil {
				return nil, err
			} else {
				que.curindex = i
			}
		}
	}
	return que, nil
}

//入队列
func (que *FileQueque) Push(data []byte) {
	if len(data) > 0 {
		que.l_data.RLock()
		defer que.l_data.RUnlock()
		if data[len(data)-1] != '\n' {
			data = append(data, '\n') //队列以换行符作为分隔符
		}

		que.f_w_data.Write(data)
	}
}

//出队列
func (que *FileQueque) Pop() ([]byte, error) {
	que.rlock.Lock()
	defer que.rlock.Unlock()
	if _, err := que.f_r_data.Seek(que.curindex, io.SeekStart); err != nil {
		return nil, err
	}
	var reader = bufio.NewReader(que.f_r_data)
	if bytes, err := reader.ReadBytes('\n'); err != nil {
		return nil, err
	} else {
		var newidx = que.curindex + int64(len(bytes))
		if err := que.writenewindex(newidx); err != nil {
			return nil, err
		} else {
			que.curindex = newidx
		}
		return bytes[:len(bytes)-1], nil
	}
}

//写新文件索引
func (que *FileQueque) writenewindex(newindex int64) error {
	if err := que.f_data_idx.Truncate(0); err != nil {
		return err
	} else {
		if _, err := que.f_data_idx.Seek(0, io.SeekStart); err != nil {
			return err
		}
		sindex := strconv.FormatInt(newindex, 10)
		if _, err := que.f_data_idx.WriteString(sindex); err != nil {
			return err
		}
	}
	return nil
}
