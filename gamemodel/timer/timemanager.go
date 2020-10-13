package timer

import (
	"time"
)

//时间管理
type TimeManager struct {
	timers []*Timer
	nowfun func() time.Time //获取当前时间的方法
}

//增加定时器
func (tm *TimeManager) _addtimer(add *Timer) *Timer {
	tm.timers = append(tm.timers, add)
	return add
}

//新建一个时间管理器
func NewTimerManager(nowf func() time.Time) *TimeManager {
	if nowf == nil {
		nowf = time.Now
	}
	return &TimeManager{nowfun: nowf}
}

/**
	固定时间间隔执行的方法
	delay int64 延迟的间隔时间 单位毫秒
	func func(*Timer,...interface{}) 执行的方法体
	params... 执行的参数
**/
func (tm *TimeManager) Loop(delay int64, fun func(*Timer, ...interface{}), params ...interface{}) *Timer {
	if delay == 0 {
		delay = 1
	}

	if fun == nil {
		return nil
	}

	return tm._addtimer(_newtimer(tm.nowfun(), delay, fun, params))
}

/**
	延迟delay毫秒后执行一次函数体
	delay int64 延迟的间隔时间 单位毫秒
	func func(...interface{}) 执行的方法体
	params... 执行的参数
**/
func (tm *TimeManager) Once(delay int64, fun func(...interface{}), params ...interface{}) *Timer {
	return tm.Loop(delay, func(timer *Timer, params ...interface{}) {
		fun(params...)
		tm.Clear(timer)
	}, params...)
}

//清理定时器 传入要清理的定时器
func (tm *TimeManager) Clear(del *Timer) {
	del.boclear = true
}

/*
	定时器主逻辑入口
	这里需要注意的是 首先Loop和Once方法的delay参数在小于Update方法帧频的时候会以Update方法的实际帧频为准
	其次目前的算法 delay参数的实际误差会在(0,(实际帧率+Update单帧执行时长)] 区间内
	因而如果希望获得精度表现更佳的定时器需要寻求其他实现方案
*/
func (tm *TimeManager) Update() {
	var now = tm.nowfun()
	var haveclear = false
	for _, timer := range tm.timers {
		if timer.boclear {
			haveclear = true
			continue
		}
		var sub = now.Sub(timer.nextt)
		if sub >= 0 { //时间已过
			timer.fun(timer, timer.Params...)
			if timer.boclear {
				haveclear = true
			} else {
				timer.nextt = now.Add(time.Millisecond * time.Duration(timer.delay)) //记录下次执行时间
			}
		}
	}
	if haveclear { //有待清理的定时器
		for i := 0; i < len(tm.timers); i++ {
			if tm.timers[i].boclear {
				tm.timers = append(tm.timers[:i], tm.timers[i+1:]...)
				i--
			}
		}
	}
}
