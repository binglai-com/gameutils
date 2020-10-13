package svrconf

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	gamebaseversion struct {
		Version int
		Data    string
	}
)

type GameBaseLoader struct {
	base     map[string][]byte //压缩目录名对应的文件数据
	address  string
	version  string
	dataname string
}

func httpgetdata(url string) (ret []byte, err error) {
	resp, gerr := http.Get(url)
	if gerr != nil {
		err = gerr
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("httpget get error: " + string(resp.StatusCode) + " resp.StatusCode != 200 url: " + url)
		return
	}

	ret, err = ioutil.ReadAll(resp.Body)
	return
}

func (l *GameBaseLoader) Init(address string, version string) error {
	l.base = make(map[string][]byte)
	l.address = address

	if version == "" { //未指定版本号 获取最新的版本号
		body, gerr := httpgetdata(l.address + "server_version")
		if gerr != nil {
			return gerr
		}

		var tmpversion gamebaseversion

		jerr := json.Unmarshal(body, &tmpversion)
		if jerr != nil {
			return jerr
		}

		l.version = fmt.Sprintf("%d", tmpversion.Version)
		l.dataname = tmpversion.Data
	} else {
		l.version = version
		l.dataname = "server_data.sg"
	}

	dataname := fmt.Sprintf("%s", l.dataname)
	body, gerr := httpgetdata(fmt.Sprintf("%s%s", l.address, dataname))
	if gerr != nil {
		return gerr
	}

	if len(body) == 0 {
		return errors.New("GameBaseLoader Init error:" + fmt.Sprintf("%s%s", l.address, dataname))
	}

	r := bytes.NewReader(body)
	zipreader, err := zip.NewReader(r, int64(len(body)))
	if err != nil {
		return errors.New("zip.NerReader err : " + err.Error())
	}

	for _, file := range zipreader.File {
		rc, err := file.Open()
		if err != nil {
			return errors.New("open zip file error :" + err.Error())
		}
		if file.Name == "data" {
			continue
		}

		buf, er := ioutil.ReadAll(rc)
		if er != nil {
			return er
		}

		l.base[file.Name] = buf
		er2 := rc.Close()
		if er2 != nil {
			return er2
		}
	}
	return nil
}

func (l *GameBaseLoader) LoadConfigFile(filename string, data interface{}) error {
	if buf, ok := l.base[filename]; !ok {
		return errors.New("load configfile openfile error , " + filename + " not found.")
	} else {
		err := json.Unmarshal(buf, data)
		if err != nil {
			return errors.New("load configfile json error : " + err.Error() + " file: " + filename)
		}

		return nil
	}

}

//清理缓存数据
func (l *GameBaseLoader) Clear() {
	l.base = make(map[string][]byte)
}

//获取所有配置表内容
func (l *GameBaseLoader) AllContent() map[string][]byte {
	return l.base
}
