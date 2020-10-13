package timer

import (
	"time"
)

//计时器
type Timer struct {
	nextt   time.Time                    //下次执行的时间
	fun     func(*Timer, ...interface{}) //执行方法体
	Params  []interface{}                //执行参数
	delay   int64                        //执行间隔 单位毫秒
	boclear bool                         //是否被移除了
}

//新建一个定时器
func _newtimer(now time.Time, delay int64, fun func(*Timer, ...interface{}), params []interface{}) *Timer {
	return &Timer{
		now.Add(time.Millisecond * time.Duration(delay)),
		fun,
		params,
		delay,
		false}
}
