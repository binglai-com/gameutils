package server

import (
	"fmt"
	"sync"
	//	"fmt"
	"net"
	"strings"
	"time"

	"gameutils/gamelog/filelog"
	"gameutils/gamenet"
)

//发给客户端的数据包
type MsgOutPacket struct {
	Cmd  uint16      //命令号
	Data interface{} //消息体
}

//网络客户端 NetClient
type NetClient struct {
	Id         uint32
	c          net.Conn
	r          bool
	w          bool
	RemoteAddr string
	RequestURI string
	alivetime  int64
	loc        sync.Mutex //保护发送队列
	msgout     [][]byte   //发送的消息

	Start      time.Time   //创建时间
	FlowIn     int         //收包大小
	FlowOut    int         //发包大小
	User       interface{} //登录的用户
	LoginState uint8       //登录状态
}

func (c *NetClient) GetId() uint32 {
	return c.Id
}

func (c *NetClient) GetRemoteAddress() string {
	return c.c.RemoteAddr().String()
}

func (c *NetClient) GetRemoteIP() string {
	address := c.GetRemoteAddress()
	n := strings.Index(address, ":")
	if n == -1 {
		return address
	} else {
		return address[:n]
	}
}

func (c *NetClient) Close() {
	c.r = false
	c.w = false
	c.c.Close()
}

func (c *NetClient) Available() bool {
	return c.w
}

func (c *NetClient) AliveTime() int64 {
	return c.alivetime
}

func (c *NetClient) GetFlowIn() int {
	return c.FlowIn
}
func (c *NetClient) GetFlowOut() int {
	return c.FlowOut
}

func GetRandomData(nlen int) []byte {
	b_buf := make([]byte, nlen)
	for i := 0; i < nlen; i++ {
		b_buf[i] = 'a'
	}
	return b_buf
}

//func (t *NetClient) SendPacketByCmd(cmd uint16, v interface{}) {
//	filelog.DEBUG("sendmsg", "session:", t.GetId(), "code:", cmd, " data:", v)
//	data, er := msgencoder.Encode(v)
//	if er != nil {
//		filelog.ERROR("netserver", fmt.Sprintf("msgencoder error: %s", er.Error()))
//		return
//	}
//	cmddata, _ := gamenet.IntToByte(cmd)
//	data = append(cmddata, data...)
//	t.SendPacket(data)
//}

func (t *NetClient) SendGameMsg(msg GameMsg) {
	binmsg := Convert2BingLaiBinMsg(msg)
	t.SendPacket(binmsg.Data)
}

//func (t *NetClient) SendMsgBufByCmd(cmd uint16, buff []byte) {
//	cmddata, _ := gamenet.IntToByte(cmd)
//	cmddata = append(cmddata, buff...)
//	t.SendPacket(cmddata)
//}

func SendGameMsg(conn net.Conn, msg GameMsg) error {
	binmsg := Convert2BingLaiBinMsg(msg)

	if wn, err := SendPacket(conn, binmsg.Data); err != nil {
		return err
	} else {
		if wn != len(binmsg.Data) {
			return fmt.Errorf("write net data wn != len(sd)")
		}
	}
	return nil
}

func SendPacket(conn net.Conn, data []byte) (int, error) {
	return conn.Write(AppendMsgHeader(data))
}

//为tcp消息体追加消息头
func AppendMsgHeader(data []byte) (res []byte) {
	if len(data) > 65530 { //大包则增加四位进行验证
		BigPacketMark, _ := gamenet.IntToByte(uint16(65535)) //65535为标识，代表大包
		allLen, _ := gamenet.IntToByte(uint32(len(data)))
		data1 := append(BigPacketMark, allLen...)
		res = append(data1, data...)
	} else {
		allLen1, _ := gamenet.IntToByte(uint16(len(data)))
		res = append(allLen1, data...)
	}
	return res
}

type ErrTimeout interface {
	Timeout() bool
	Error() string
}
type PacketReader func(deadline time.Time) (error, GameMsg)
type MsgHandler func(cmd uint16, msg GameMsg) error //消息处理

//读取收到的下一条消息
func (reader PacketReader) Next(deadline time.Time) (error, GameMsg) {
	return reader(deadline)
}

//等待指定的消息号
func (reader PacketReader) Wait(cmd uint16, deadline time.Time, handler MsgHandler) (error, GameMsg) {
	for {
		if err, msg := reader.Next(deadline); err != nil {
			if timeouterr, ok := err.(ErrTimeout); ok && timeouterr.Timeout() {
				return fmt.Errorf("wait cmd %d time out", cmd), nil //已超时
			} else {
				return err, msg
			}
		} else {
			var msgcmd = msg.Cmd()
			if msgcmd == cmd {
				return nil, msg
			} else {
				if handler != nil {
					if err := handler(msgcmd, msg); err != nil {
						return err, msg
					}
				}
			}
		}
	}
}

//等待任意的消息号
func (reader PacketReader) WaitAny(cmds []uint16, deadline time.Time) (error, GameMsg) {
	for {
		if err, msg := reader.Next(deadline); err != nil {
			if timeouterr, ok := err.(ErrTimeout); ok && timeouterr.Timeout() {
				return fmt.Errorf("waitany cmd %v time out", cmds), nil //已超时
			} else {
				return err, msg
			}
		} else {
			var bofind = false
			var code = msg.Cmd()
			for _, v := range cmds {
				if v == code {
					bofind = true
					break
				}
			}

			if bofind {
				return nil, msg
			}
		}
	}
}

//等待所有的消息号
func (reader PacketReader) WaitAll(cmds []uint16, deadline time.Time) (error, []GameMsg) {
	var ret = make([]GameMsg, 0)
	if len(cmds) == 0 {
		return nil, ret
	}
	var dict = make(map[uint16]uint8)
	for _, v := range cmds {
		dict[v] = 1
	}

	for {
		if err, msg := reader.Next(deadline); err != nil {
			if timeouterr, ok := err.(ErrTimeout); ok && timeouterr.Timeout() {
				return fmt.Errorf("waitall cmd %v time out", cmds), nil //已超时
			} else {
				return err, ret
			}
		} else {
			var code = msg.Cmd()
			for _, v := range cmds {
				if v == code {
					ret = append(ret, msg)
					delete(dict, code)
					break
				}
			}

			if len(dict) == 0 {
				return nil, ret
			}
		}
	}
}

func CreatePacketReader(c net.Conn) PacketReader {
	packet := make([]byte, 0)
	buf := make([]byte, 1024)
	return func(deadline time.Time) (error, GameMsg) {
		var havedeadline = !deadline.IsZero()
		for {
			if havedeadline {
				c.SetReadDeadline(deadline)
			}
			rlen, err := c.Read(buf)

			if havedeadline {
				c.SetReadDeadline(time.Time{})
			}
			if err != nil {
				if errtimeout, ok := err.(ErrTimeout); ok && errtimeout.Timeout() {
					if rlen > 0 {
						panic("read from conn err timeout with rlen > 0")
					}
				}

				return err, nil
			}
			packet = append(packet, buf[:rlen]...)
			for len(packet) > 3 { ///2个字节包长度 2个字节的命令号
				hlen := gamenet.Uint16(packet[0:2])
				start := 2
				readlen := int(hlen) + 2
				if hlen < 2 {
					c.Close()
					return fmt.Errorf("hlen err datalen < 1, data : %s", string(packet[2:])), nil
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
					if len(packet) == readlen {
						packet = packet[:0]
					} else {
						packet = packet[readlen:]
					}

					return nil, data
				} else {
					break
				}
			}
		}
	}
}

func (c *NetClient) SendPacket(data []byte) {
	c.Send(AppendMsgHeader(data))
}

func (c *NetClient) Send(data []byte) {
	if c.c == nil || !c.w {
		filelog.ERROR("netserver", "netclient no connect to send")
		return
	}
	//	c.loc.Lock()
	//	c.msgout = append(c.msgout, data)
	//	c.loc.Unlock()
	c.c.Write(data)
}
