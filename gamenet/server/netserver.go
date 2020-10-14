package server

import (
	"io"
	"net"
	"runtime/debug"
	"time"

	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/binglai-com/gameutils/gamenet"

	"golang.org/x/net/websocket"
)

type NetServer struct {
	l        net.Listener
	h        func(data GameMsg, c *NetClient)
	hdis     func(*NetClient)
	hconnect func(*NetClient)
	cs       []*NetClient
	freesess chan uint32
	timeout  int64
	isrun    bool //是否还在运行
}

func (s *NetServer) Close() {
	s.isrun = false
	if s.l != nil {
		s.l.Close()
	}
}

///设置超时默认为0
func (s *NetServer) SetTimeOut(time int64) {
	s.timeout = time
}

func (this *NetServer) DelSess(sid uint32) {
	client := this.GetClient(sid)
	if client == nil {
		return
	}

	this.cs[sid] = nil
	this.freesess <- sid
}

//删除Client
func (this *NetServer) DelClient(sid uint32) {
	client := this.GetClient(sid)
	if client != nil {
		this.DelSess(client.GetId())
		client.Close()
	}
}

//删除所有client
func (this *NetServer) DelAllClient() {
	for i := 1; i < len(this.cs); i++ {
		this.DelClient(uint32(i))
	}
}

//获取空闲的会话id
func (this *NetServer) GetFreeSess() uint32 {
	return <-this.freesess
}

func (s *NetServer) acceptclient(c *NetClient) {
	defer c.Close()
	defer func() {
		if err := recover(); err != nil {
			filelog.ERROR("gocrash", "netserver accept over:", err, "Sid:", c.Id, ",debug:", string(debug.Stack()))
		}
	}()

	go func() { //网络写线程
		var senddata *[]byte = nil
		for c.w && c.c != nil {
			c.loc.Lock()
			if len(c.msgout) > 0 {
				senddata = &(c.msgout[0])
				c.msgout = c.msgout[1:]
			}
			c.loc.Unlock()

			if senddata != nil {
				c.alivetime = time.Now().Unix()
				sendlen := len(*senddata)
				n, err := c.c.Write(*senddata)
				if err != nil && c.w && c.c != nil {
					filelog.ERROR("netserver", "netclient write err:", err.Error())
					c.Close()
					return
				}
				if n != sendlen && c.w && c.c != nil {
					filelog.ERROR("netserver", "netclient write err: send len:", n, ",expected sendlen", sendlen)
					c.Close()
					return
				}

				//发送成功
				senddata = nil
				c.FlowOut += n
				time.Sleep(time.Millisecond)
			} else { //队列空闲
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	if s.hconnect != nil {
		s.hconnect(c)
	}
	packet := make([]byte, 0)
	buf := make([]byte, 1024)
	for c.r && c.c != nil {
		rlen, err := c.c.Read(buf)
		if err != nil && c.r && c.c != nil {
			if err != io.EOF {
				filelog.ERROR("netserver", "netclient read err:", err.Error(), " c.Id : ", c.Id)
			} else {
				filelog.INFO("netserver", "netclient read EOF, c.Id : ", c.Id)
			}
			break
		}
		c.FlowIn += rlen
		///解析包
		c.alivetime = time.Now().Unix()
		packet = append(packet, buf[:rlen]...)
		for len(packet) > 3 { ///2个字节包长度 2个字节的命令号
			hlen := gamenet.Uint16(packet[0:2])
			start := 2
			readlen := int(hlen) + 2
			if hlen < 2 {
				filelog.ERROR("netserver", "hlen error:", hlen, " datalen:", len(packet[2:]), " data:", string(packet[2:]))
			}
			if hlen == 65535 { //大包
				if len(packet) < 6 {
					break
				}
				bigLen := gamenet.Uint32(packet[2:6])
				start = 6
				readlen = int(bigLen) + 6
			}
			if len(packet) >= readlen {
				data := new(BingLaiBinMsg)
				data.Data = packet[start:readlen]
				s.h(data, c)
				if len(packet) == readlen {
					packet = make([]byte, 0)
				} else {
					packet = packet[readlen:]
				}
			} else {
				break
			}
		}
	}
	if s.hdis != nil {
		s.hdis(c)
	}
	c.r = false
	c.w = false
	s.DelSess(c.Id)
}

func ServerAccept(s *NetServer) {
	defer s.l.Close()
	defer func() {
		if err := recover(); err != nil {
			filelog.ERROR("gocrash", "ServerAccept over error:", err, ",debug:", string(debug.Stack()))
		}
		filelog.INFO("netserver", "tcp server accept shutdown.")
	}()
	for s.isrun {
		cconn, cerr := s.l.Accept()
		if cerr != nil && s.isrun {
			filelog.ERROR("netserver", "ServerAccept over:", cerr.Error())
			return
		}

		if cconn == nil {
			continue
		}

		sessid := s.GetFreeSess() //获取可用的会话id
		c := new(NetClient)
		c.c = cconn
		c.r = true
		c.w = true
		c.Id = sessid
		c.RemoteAddr = c.GetRemoteAddress()
		s.cs[c.Id] = c
		c.alivetime = time.Now().Unix()
		c.Start = time.Now()

		go s.acceptclient(c)
	}
}

func (c *NetServer) GetClient(sid uint32) *NetClient {
	if sid == 0 || int(sid) >= len(c.cs) {
		return nil
	}
	return c.cs[sid]
}

func (c *NetServer) GetCount() int {
	return len(c.cs) - 1 - len(c.freesess)
}

//WebSocket
func (s *NetServer) WebSocketHandler(wsconn *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			filelog.ERROR("gocrash", "WebSocketHandler error:", err)
			filelog.ERROR("gocrash", "Stack:", string(debug.Stack()))
		}
	}()

	if wsconn == nil {
		return
	}

	filelog.DEBUG("netserver", "WebSocketHandler ", wsconn.IsClientConn(), wsconn.IsServerConn(), wsconn.Request().Method, wsconn.Request().URL, wsconn.Request().RequestURI)

	sessid := s.GetFreeSess()

	c := new(NetClient)
	wsconn.PayloadType = websocket.BinaryFrame //收发二进制数据
	c.c = wsconn
	c.r = true
	c.w = true
	c.Id = sessid
	c.RemoteAddr = wsconn.Request().RemoteAddr
	c.RequestURI = wsconn.Request().RequestURI
	s.cs[c.Id] = c
	c.alivetime = time.Now().Unix()
	c.Start = time.Now()

	//循环接收消息
	s.acceptclient(c)
}

var MAXONLINE = uint32(20000)

///创建Server
func NewNetServer(tcpaddress string, handler func(data GameMsg, c *NetClient), onConnect func(*NetClient), onDisconnect func(*NetClient)) (*NetServer, error) {
	t := new(NetServer)
	t.isrun = true
	t.h = handler
	t.hconnect = onConnect
	t.hdis = onDisconnect
	t.cs = make([]*NetClient, MAXONLINE)
	t.freesess = make(chan uint32, MAXONLINE)
	for i := uint32(1); i < MAXONLINE; i++ {
		t.freesess <- i
	}
	t.timeout = 0

	if tcpaddress != "" {
		listen, lerr := net.Listen("tcp", tcpaddress)
		if lerr != nil {
			return nil, lerr
		}
		t.l = listen
		go ServerAccept(t)
	}

	go ServerDaemon(t)
	return t, nil
}

func ServerDaemon(s *NetServer) {
	for {
		defer func() {
			if err := recover(); err != nil {
				filelog.ERROR("gocrash", "ServerDaemon over error:", err, ",debug:", string(debug.Stack()))
			}
		}()
		if s.timeout > 0 {
			now := time.Now().Unix()
			filelog.INFO("netserver", "serverclients cnt : ", s.GetCount())
			for _, v := range s.cs {
				if v != nil && now-v.AliveTime() > s.timeout {
					v.Close()
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}
