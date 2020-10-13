package schedule

import (
	"time"
)

//获取某个时间点之后的下一个的时间点
func GetNextDayTime(t time.Time, hour, min, sec int) time.Time {
	var t2 = time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, t.Location())
	if t.After(t2) {
		return t2.Add(time.Hour * 24)
	} else {
		return t2
	}
}
