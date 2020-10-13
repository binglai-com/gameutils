package gamerpc

import (
	"fmt"
	"testing"
	"time"
)

type (
	TestRpcParam1 struct {
		A string
	}

	TestRpcRes struct {
		B string
	}
)

func Test_Call(t *testing.T) {
	//	_startrpcservice(":5001", t)
	time.Sleep(time.Second * 2)

	for i := 0; i < 2; i++ {
		var res TestRpcRes
		err := Call(":5001", "TestRpcSvr.Rpc1Error", TestRpcParam1{"ping"}, &res)
		fmt.Println(res.B)
		if err != nil {
			t.Log(err.Error())
		}
	}

	if len(defaultpool._clients) != 1 {
		t.Fatal("got more clients than expected. not clients : ", len(defaultpool._clients))
	}

	fmt.Println("waiting service to restart")
	//	time.Sleep(60 * time.Second)

	//	_startrpcservice(":5001")
	fmt.Println("call rpc1!!")
	var res TestRpcRes
	err := Call(":5001", "TestRpcSvr.Rpc1", TestRpcParam1{"ping"}, &res)
	if err != nil {
		t.Fatal("rpc call failed after rpc service restart. err : ", err.Error())
	}
}
