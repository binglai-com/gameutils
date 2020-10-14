package gamerpc

import (
	"bufio"
	"encoding/gob"
	"log"
	"net/http"
	"net/rpc"
	"runtime/debug"

	"github.com/binglai-com/gameutils/gamelog/filelog"
)

//rpc请求的handler
func rpchandler(rsp http.ResponseWriter, req *http.Request) {
	buf := bufio.NewWriter(rsp)
	var servercodec = &HttpRpcSvrCodec{
		gob.NewDecoder(req.Body),
		gob.NewEncoder(buf),
		buf,
		false,
	}

	//	var rpcstart = time.Now()
	defer func() {
		if err := recover(); err != nil { //uncaught exception
			filelog.ERROR("gocrash", "rpcserver Process got err : ", err, " stack : ", string(debug.Stack()))
		}
	}()
	err := rpc.DefaultServer.ServeRequest(servercodec)
	if err != nil {
		filelog.ERROR("gamerpc", "ServerRequest err : ", err.Error())
	}
}

type HttpRpcSvrCodec struct {
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *HttpRpcSvrCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *HttpRpcSvrCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *HttpRpcSvrCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *HttpRpcSvrCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return nil
}

func init() {
	http.HandleFunc("/_binglaiRPC_/", rpchandler)
	rpc.HandleHTTP()
	filelog.INFO("gamerpc", "init rpchandler")
}
