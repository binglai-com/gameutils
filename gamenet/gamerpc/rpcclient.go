package gamerpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"gameutils/gamelog/filelog"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

var maxconnretrynum = 1 //网络连接异常导致的最大重试次数

type clientpool struct {
	sync.RWMutex
	_clients map[string]*rpc.Client //依据地址存储的rpcclient map[rpcaddr]*rpc.Client
}

var defaultpool = &clientpool{_clients: make(map[string]*rpc.Client)}

func (this *clientpool) _getclient(addr string) *rpc.Client {
	this.RLock()
	c := this._clients[addr]
	this.RUnlock()
	if c == nil {
		if tmpc, err := rpc.DialHTTP("tcp", addr); err != nil {
			filelog.ERROR("gamerpc", "dial http ", addr, " err : ", err.Error())
			return nil
		} else {
			c = tmpc
			this.Lock()
			this._clients[addr] = c
			this.Unlock()
		}
	}
	return c
}

//关闭rpc连接
func (this *clientpool) _closeclient(addr string, c *rpc.Client) {
	c.Close() //清理该调用
	this.Lock()
	tmp := this._clients[addr]
	if tmp == c {
		delete(this._clients, addr)
	}
	this.Unlock()
}

//调用
func (this *clientpool) _call(addr string, servicename string, args interface{}, res interface{}, retrynum int) (err error) {
	var boretry = false
	c := this._getclient(addr)
	if c == nil {
		return fmt.Errorf("get rpc client fail, addr:%s ", addr)
	}
	err = c.Call(servicename, args, res)
	if err != nil {
		if _, ok := err.(net.Error); ok { //net package error
			boretry = true
			goto closeclient
		} else {
			switch err {
			case rpc.ErrShutdown:
				boretry = true
				goto closeclient
			case io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe, io.ErrNoProgress, io.ErrShortBuffer, io.ErrShortWrite:
				goto closeclient
			default:
				//gob error
				errmsg := err.Error()
				msglen := len(errmsg)
				if msglen > 3 && errmsg[:3] == "gob" {
					goto closeclient
				} else if msglen >= 23 && errmsg[:10] == "value for " && errmsg[msglen-13:] == " out of range" {
					goto closeclient
				} else if errmsg == "invalid message length" || errmsg == "extra data in buffer" {
					goto closeclient
				} else if msglen >= 39 && errmsg[:39] == "can't represent recursive pointer type " {
					goto closeclient
				}
			}
		}
	}

	return
closeclient:
	this._closeclient(addr, c) //关闭连接
	if boretry && retrynum < maxconnretrynum {
		err = this._call(addr, servicename, args, res, retrynum+1)
	}
	return
}

//Rpc调用 沿用rpc标准库中的机制复用rpcclient复用conn 目前测试的几种机制中，性能最好，但是不易处理请求应答的超时问题
func Call(addr string, servicename string, args interface{}, reply interface{}) error {
	return defaultpool._call(addr, servicename, args, reply, 0)
}

var defaulthttpclient = &http.Client{Timeout: time.Minute} //rpc调用超时时间定义为1分钟

//Rpc调用1 基于http post请求的rpc封装，conn复用策略基于所使用的http版本，请求的超时机制同httpclient的超时机制，在httpserver支持keepalive（复用conn）的情形下性能较好
func Call1(addr string, servicename string, args interface{}, reply interface{}) error {
	call := new(rpc.Call)
	call.ServiceMethod = servicename
	call.Args = args
	call.Reply = reply

	var req = new(rpc.Request)
	req.ServiceMethod = call.ServiceMethod
	req.Seq = 1

	pipereader, pipewriter := io.Pipe()
	encBuf := bufio.NewWriter(pipewriter)
	var gobencoder = gob.NewEncoder(encBuf)
	go func() {
		if err := gobencoder.Encode(&req); err != nil {
			pipewriter.CloseWithError(err)
			return
		}
		if err := gobencoder.Encode(call.Args); err != nil {
			pipewriter.CloseWithError(err)
		}

		encBuf.Flush()
		pipewriter.Close()
	}()
	resp, err := defaulthttpclient.Post(fmt.Sprintf("http://%s/_binglaiRPC_/", addr), "application/octet-stream", pipereader)
	pipereader.Close()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var dec = gob.NewDecoder(resp.Body)

	response := rpc.Response{}
	err = dec.Decode(&response)
	if err != nil {
		return err
	}
	switch {
	case response.Error != "":
		// We've got an error response. Give this to the request;
		// any subsequent requests will get the ReadResponseBody
		// error if there is one.
		call.Error = rpc.ServerError(response.Error)
		err = dec.Decode(nil)
		if err != nil {
			err = errors.New("reading error body: " + err.Error())
			return err
		}
		return call.Error
	default:
		err = dec.Decode(call.Reply)
		if err != nil {
			err = errors.New("reading body " + err.Error())
			return err
		}
	}
	return nil
}

//Rpc调用2 使用rpc标准库中的机制 但是不复用rpcclient和conn 方便针对一次请求做超时管理但是性能较差
func Call2(addr string, servicename string, args interface{}, reply interface{}) error {
	if tmpc, err := rpc.DialHTTP("tcp", addr); err != nil {
		filelog.ERROR("gamerpc", "dial http ", addr, " err : ", err.Error())
		return err
	} else {
		err = tmpc.Call(servicename, args, reply)
		tmpc.Close()
		return err
	}
}
