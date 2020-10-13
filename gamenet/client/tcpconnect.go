package client

import (
	"gameutils/gamelog/filelog"
	"gameutils/gamenet"
	"gameutils/gamenet/server"
	"net"
	"sync"
	"time"
)

const (
	tcpconnectlog = "tcpconnect"
)

type Connect struct {
	c            net.Conn
	h            func(data server.GameMsg)
	r            bool
	w            bool
	ondisconnect func()

	loc    sync.Mutex
	msgout [][]byte
}

func TcpConnectWork(t *Connect) {
	//创建写线程
	go func() {
		var senddata *[]byte = nil
		var alivedur = 5 * time.Minute
		var nextalive = time.Now().Add(alivedur)
		var alivemsg = server.Convert2BingLaiBinMsg(&server.GameChanMsg{65535, []byte{}, nil, nil})
		for t.w && t.c != nil {
			t.loc.Lock()
			if len(t.msgout) > 0 {
				senddata = &(t.msgout[0])
				t.msgout = t.msgout[1:]
			}
			t.loc.Unlock()
			if senddata != nil {
				sendlen := len(*senddata)
				n, err := t.c.Write(*senddata)
				if err != nil && t.w && t.c != nil {
					filelog.ERROR("netserver", "netclient write err:", err.Error())
					t.Close()
					return
				}
				if n != sendlen && t.w && t.c != nil {
					filelog.ERROR("netserver", "netclient write err: send len:", n, ",expected sendlen", sendlen)
					t.Close()
					return
				}

				//发送成功
				senddata = nil
				nextalive = time.Now().Add(alivedur)
				time.Sleep(time.Millisecond)
			} else { //队列空闲
				now := time.Now()
				if now.After(nextalive) {
					nextalive = now.Add(alivedur)
					sendlen := len(alivemsg.Data)
					n, err := t.c.Write(alivemsg.Data)
					if err != nil && t.w && t.c != nil {
						filelog.ERROR("netserver", "netclient write err:", err.Error())
						t.Close()
						return
					}
					if n != sendlen && t.w && t.c != nil {
						filelog.ERROR("netserver", "netclient write err: send len:", n, ",expected sendlen", sendlen)
						t.Close()
						return
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	packet := make([]byte, 0)
	for t.r && t.c != nil {
		buf := make([]byte, 1024)
		rlen, rer := t.c.Read(buf)
		if rer != nil && t.r && t.c != nil {
			filelog.ERROR("tcpconnect", "read from conn fail : ", rer.Error())
			break
		}
		buf = buf[:rlen]
		///解析包
		packet = append(packet, buf...)
		for len(packet) > 3 { ///2个字节包长度 2个字节的命令号,若大包，前2个字节标志位65535，中间4个字节长度，接着2个字节的命令号
			hlen := gamenet.Uint16(packet[0:2])
			start := 2
			readlen := int(hlen) + 2
			if hlen == 65535 {
				if len(packet) < 6 {
					break
				}
				bigLen := gamenet.Uint32(packet[2:6])
				start = 6
				readlen = int(bigLen) + 6
			}
			if len(packet) >= readlen {
				msg := new(server.BingLaiBinMsg)
				msg.Data = packet[start:readlen]
				t.h(msg)
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

	if t.ondisconnect != nil {
		t.ondisconnect()
	}

	t.r = false
	t.w = false
	t.Close()
}

///创建一个tcp链接
func NewTcpConnect(address string, handler func(server.GameMsg), ondisconnecthandler func()) *Connect {
	t := new(Connect)
	t.h = handler
	c, er := net.Dial("tcp", address)
	if er != nil {
		filelog.ERROR("tcpconnect", "net.dial error:", er.Error())
		t.r = false
		t.w = false
		t.c = nil
		return nil
	} else {
		t.r = true
		t.w = true
		t.c = c
	}

	t.ondisconnect = ondisconnecthandler
	go TcpConnectWork(t)
	return t
}

//根据conn创建一个Connect
func CreateConnect(c net.Conn, handler func(server.GameMsg), ondisconnecthandler func()) *Connect {
	t := new(Connect)
	t.h = handler
	t.r = true
	t.w = true
	t.c = c
	t.ondisconnect = ondisconnecthandler
	go TcpConnectWork(t)

	return t
}

func (t *Connect) IsAlive() bool {
	return t.w
}

func (t *Connect) SendGameMsg(msg server.GameMsg) {
	binmsg := server.Convert2BingLaiBinMsg(msg)
	t.SendPacket(binmsg.Data)
}

func (c *Connect) SendPacket(data []byte) {
	if len(data) > 65530 { //大包则增加四位进行验证
		BigPacketMark, _ := gamenet.IntToByte(uint16(65535)) //65535为标识，代表大包
		allLen, _ := gamenet.IntToByte(uint32(len(data)))
		data1 := append(BigPacketMark, allLen...)
		data = append(data1, data...)
	} else {
		allLen1, _ := gamenet.IntToByte(uint16(len(data)))
		data = append(allLen1, data...)
	}
	c.Send(data)
}

func (c *Connect) Send(data []byte) {
	if c.c == nil || !c.w {
		filelog.ERROR("tcpconnect", "no connect to send")
		return
	}

	c.loc.Lock()
	c.msgout = append(c.msgout, data)
	c.loc.Unlock()
}

///关闭链接
func (t *Connect) Close() {
	t.r = false
	t.w = false
	if t.c != nil {
		t.c.Close()
	}
}
