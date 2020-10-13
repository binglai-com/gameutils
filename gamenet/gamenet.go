package gamenet

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"
)

//uint32类型数字
func Uint32(d []byte) uint32 {
	var v uint32 = 0
	header := bytes.NewBuffer(d)
	binary.Read(header, binary.BigEndian, &v)
	return v
}

//uint16类型数字
func Uint16(d []byte) uint16 {
	var v uint16 = 0
	header := bytes.NewBuffer(d)
	binary.Read(header, binary.BigEndian, &v)
	return v
}

//4字节数字转字节
func IntToByte(v interface{}) ([]byte, error) {
	header := bytes.NewBuffer([]byte{})
	err := binary.Write(header, binary.BigEndian, v)
	if err != nil {
		return []byte{}, err
	}
	return header.Bytes(), nil
}

func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

//进行zlib解压缩
func DoZlibUnCompress(compressSrc []byte) []byte {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, er := zlib.NewReader(b)
	if er != nil {
		//PrintLog("DoZlibUnCompress error:", er.Error(), " data:", string(compressSrc))
		return nil
	}
	io.Copy(&out, r)
	return out.Bytes()
}

//获取内网IP
func GetInternal() ([]string, error) {
	var res []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, v := range addrs {
		if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				res = append(res, ipnet.IP.String())
			}
		}
	}
	if len(res) != 0 {
		return res, nil
	}

	return nil, errors.New("not find internal IP")
}

//获取外网IP
func GetExternal() (string, error) {
	httpclient := http.Client{Timeout: 2 * time.Second}
	resp, err := httpclient.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return "", err2
	}
	r, err3 := regexp.Compile("\\s")
	if err3 != nil {
		return "", err3
	}
	res = r.ReplaceAll(res, []byte(""))
	return string(res), nil
}

//获取空闲端口号
func GetFreePort(num int) ([]int, error) {
	index := 0
	res := make([]int, 0)
	for index < num {
		index++
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, err
		}
		defer l.Close()
		if v, ok := l.Addr().(*net.TCPAddr); ok {
			res = append(res, v.Port)
		} else {
			return nil, errors.New("*net.tcpaddr assertion failure.")
		}
	}
	return res, nil
}

//获取可用的链接地址
func GetAvalibleTcpAddr(addrs []string) (string, error) {
	for _, v := range addrs {
		c := make(chan error, 1)
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				conn.Close()
			}
			c <- err
		}(v)

		select {
		case e := <-c:
			if e == nil {
				return v, nil
			}
		case <-time.After(time.Millisecond * 100):
		}
	}
	return "", errors.New("no available tcp addr")
}
