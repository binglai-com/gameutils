package svrconf

import (
	"net"
	"os"

	"github.com/binglai-com/gameutils/gamelog/filelog"
)

//flash安全沙箱服务器  需要制定一个沙箱端口和一个跨域访问的策略文件
/* 策略文件示例

<?xml version="1.0"?>
<cross-domain-policy>
	<allow-access-from domain="*" to-ports="*"/>
</cross-domain-policy>
*/

//服务器goroutine
func _serverrun(l net.Listener, buf []byte) {
	for {
		c, err := l.Accept()
		if err != nil {
			filelog.INFO("box", "svr listen err : ", err.Error())
			continue
		}
		c.Write(buf)
		go func(conn net.Conn) {
			readBuf := make([]byte, 65535)
			n, err := conn.Read(readBuf)
			if err == nil {
				filelog.INFO("box", "ip:", conn.RemoteAddr().(*net.TCPAddr).IP, ",content:", string(readBuf[:n]), ",len:", n)
			} else {
				filelog.ERROR("box", "ip:", conn.RemoteAddr().(*net.TCPAddr).IP, ",error:", err.Error())
			}
			conn.Close()
		}(c)
	}
}

//初始化服务器
func InitFlashPolicySvr(addr, filepath string) error {
	listen, err1 := net.Listen("tcp", addr)
	if err1 != nil {
		return err1
	}

	file, err := os.OpenFile(filepath, os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	buf := make([]byte, 8192)
	nlen, _ := file.Read(buf)
	file.Close()
	buf = buf[:nlen]
	go _serverrun(listen, buf)
	return nil
}
