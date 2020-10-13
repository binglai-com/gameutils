package filequeue

import (
	"fmt"
	"io"
	"testing"
)

func Test_Queue(t *testing.T) {
	var que, err = InitQueue("que1")
	if err != nil {
		t.Fatalf("init que fail, err : %s", err.Error())
	}
	for i := 0; i < 10; i++ {
		que.Push([]byte(fmt.Sprintf("test%d", i)))
	}

	for {
		var data, poperr = que.Pop()
		if poperr != nil {
			if poperr == io.EOF {
				fmt.Println("read over.")
			} else {
				t.Fatalf("pop fail : %s", poperr.Error())
			}
			break
		}

		fmt.Println("pop data : ", string(data))
	}

}
