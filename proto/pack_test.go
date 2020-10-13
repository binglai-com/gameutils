/*!
 * <测试>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"log"
	"testing"
)

type TestC struct {
	J string
	H uint32
}

type TestB struct {
	F int32
	G string
	K TestC
}

type test struct {
	T TestC
	B int32
	C []TestB
	D string
	E []string
}

func Test_PackFloat(t *testing.T) {
	var data = struct {
		F32 float32
		F64 float64
	}{}
	data.F32 = float32(100.23)
	data.F64 = float64(10000000000.1213123)
	res, err := Pack(data)
	if err != nil {
		t.Fatal("Pack err : ", err.Error())
	}

	data.F32 = 0
	data.F64 = 0
	if err := UnPack(res, &data); err != nil {
		t.Fatal("Unpack err : ", err.Error())
	}

	if data.F32 != float32(100.23) {
		t.Fatal("Unpack float32 value change.")
	}

	if data.F64 != float64(10000000000.1213123) {
		t.Fatal("Unpack float64 value change.")
	}

}

//[0, 0, 0, 99, 0, 1, 0, 0, 3, 231, 119, 111, 114, 108, 100, 0, 0, 39, 102, 116, 101, 115, 116, 32, 111, 118, 101, 114, 104, 101, 108, 108, 111, 0, 3, 97, 66, 67, 68, 69]
func Test_Pack(t *testing.T) {
	var data test
	data.T.J = "测试中文字符串"
	data.B = 99
	data.D = "hello"
	for i := 0; i < 1; i++ {
		var testb TestB
		testb.F = 999
		testb.G = "world"
		var testc TestC
		testc.H = 10086
		testc.J = "test over"
		testb.K = testc
		data.C = append(data.C, testb)
	}
	data.E = []string{"a", "B", "CDE"}

	log.Println("打包前1 :", data)

	bys, err1 := Pack(&data)
	if err1 != nil {
		print("err1 :", err1.Error())
	} else {
		log.Println("打包后字节数据1 :", bys)
	}

	var unpakdata test
	UnPack(bys, &unpakdata)
	log.Println("解包后1 :", unpakdata)
}
