package server

import (
	"bytes"
	"encoding/binary"
	"gameutils/gamelog/filelog"
	"gameutils/gamenet"
	"strconv"
	"strings"
)

//游戏消息接口  网络或者golang channle中的消息
type GameMsg interface {
	Cmd() uint16
	Body() []byte
	Len() int           //获取消息号和消息体的总长度
	Res() chan error    //响应结果
	Value() interface{} //消息体对象（已经解析好的）
}

//返回结果
func SafeRes(msg GameMsg, e error) {
	if msg.Res() != nil {
		select {
		case msg.Res() <- e:
		default:
		}
	}
}

//内部管道定义的消息体
type GameChanMsg struct {
	Code uint16
	Data []byte
	Ret  chan error
	V    interface{}
}

func (m *GameChanMsg) Cmd() uint16        { return m.Code }
func (m *GameChanMsg) Body() []byte       { return m.Data }
func (m *GameChanMsg) Len() int           { return 0 }
func (m *GameChanMsg) Res() chan error    { return m.Ret }
func (m *GameChanMsg) Value() interface{} { return m.V }

//转换成兵来2进制消息
func Convert2BingLaiBinMsg(msg GameMsg) *BingLaiBinMsg {
	cmddata, _ := gamenet.IntToByte(msg.Cmd())
	cmddata = append(cmddata, msg.Body()...)
	ret := new(BingLaiBinMsg)
	ret.Data = cmddata
	return ret
}

//兵来网络二进制消息格式
type BingLaiBinMsg struct {
	Data []byte
}

func (m *BingLaiBinMsg) Cmd() uint16 {
	if len(m.Data) < 2 {
		return 0
	}

	header := bytes.NewBuffer(m.Data[0:2])
	code := uint16(0)
	binary.Read(header, binary.BigEndian, &code)
	return code
}
func (m *BingLaiBinMsg) Body() []byte {
	if len(m.Data) < 2 {
		ret := make([]byte, 0)
		return ret
	}
	return m.Data[2:]
}
func (m *BingLaiBinMsg) Len() int           { return len(m.Data) }
func (m *BingLaiBinMsg) Res() chan error    { return nil }
func (m *BingLaiBinMsg) Value() interface{} { return nil }

//兵来json网络消息
type BingLaiJsonMsg struct {
	Data string
}

func (m *BingLaiJsonMsg) Cmd() uint16 {
	pos := strings.Index(m.Data, "{")
	code := 0
	var err error
	if pos != -1 {
		code, err = strconv.Atoi(m.Data[:pos])
		if err != nil {
			filelog.ERROR("netserver", "recv error pos:", pos, " message:", m.Data)
			return 0
		}
	} else {
		code, err = strconv.Atoi(m.Data)
		if err != nil {
			filelog.ERROR("netserver", "recv error pos:", pos, " message:", m.Data)
			return 0
		}
	}
	return uint16(code)
}
func (m *BingLaiJsonMsg) Body() []byte {
	pos := strings.Index(m.Data, "{")
	var data = make([]byte, 0)
	if pos != -1 {
		data = []byte(m.Data[pos:])
	}
	return data
}
func (m *BingLaiJsonMsg) Len() int           { return len(m.Data) }
func (m *BingLaiJsonMsg) Res() chan error    { return nil }
func (m *BingLaiJsonMsg) Value() interface{} { return nil }
