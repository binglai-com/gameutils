package timer

import (
	"fmt"
	"testing"
	"time"
)

//测试loop方法
func Test_Loop(t *testing.T) {
	var tm = NewTimerManager(time.Now)

	var fun = func(timer *Timer, params ...interface{}) {
		var now = time.Now()
		var lastframetime = params[0].(time.Time)
		var last = now.Sub(lastframetime).Nanoseconds()
		if last != 1e9 {
			fmt.Println("since last out : ", last)
		}
		//		fmt.Println("since last : ", now.Sub(lastframetime).Nanoseconds())
		timer.Params[0] = now
	}

	for i := 0; i < 1; i++ {
		tm.Loop(1000, fun, time.Now())
	}

	var testend = time.Now().Add(time.Second * 10)
	for time.Now().Before(testend) {
		tm.Update()
		time.Sleep(time.Millisecond)
	}
}

//测试 once 方法
func Test_Once(t *testing.T) {
	var tm = NewTimerManager(time.Now)

	var last = time.Now()
	tm.Once(1000, func(params ...interface{}) {
		var cost = time.Since(last)
		if cost != 1e9 {
			fmt.Println("real cost : ", cost)
		}
	})

	t.Log("total timers : ", len(tm.timers))

	var testend = time.Now().Add(time.Second * 10)
	for time.Now().Before(testend) {
		tm.Update()
		time.Sleep(time.Millisecond)
	}

	t.Log("timers left : ", len(tm.timers))
}
