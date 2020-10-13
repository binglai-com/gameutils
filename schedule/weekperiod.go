package schedule

import (
	"time"
)

//获取前一个 星期wd 对应的时间
func GetPreWeekDayTime(t time.Time, wd int, hour, min, sec int) time.Time {
	if wd >= 7 {
		wd = 0
	} else {
		if wd < 1 {
			wd = 1
		}
	}

	if int(t.Weekday()) == wd {
		//判断今天
		tar := time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, time.Local)
		if t.Before(tar) { //时间已过 算下一周期的指定时间
			t = t.AddDate(0, 0, -1)
		} else {
			return tar
		}
	} else {
		t = t.AddDate(0, 0, -1)
	}

	for i := 0; i < 14; i++ {
		if int(t.Weekday()) == wd {
			return time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, time.Local)
		}
		t = t.AddDate(0, 0, -1)
	}

	return time.Time{}
}

//获取后一个 星期wd 对应的时间
func GetNextWeekDayTime(t time.Time, wd int, hour, min, sec int) time.Time {
	if wd >= 7 {
		wd = 0
	} else {
		if wd < 1 {
			wd = 1
		}
	}

	if int(t.Weekday()) == wd {
		//判断今天
		tar := time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, time.Local)
		if t.After(tar) { //时间已过 算下一周期的指定时间
			t = t.AddDate(0, 0, 1)
		} else {
			return tar
		}
	} else {
		t = t.AddDate(0, 0, 1)
	}

	for i := 0; i < 14; i++ {
		if int(t.Weekday()) == wd {
			return time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, time.Local)
		}
		t = t.AddDate(0, 0, 1)
	}
	return time.Time{}
}

//游戏中以周为周期描述时间段
type WeekPeriod []int

func (p WeekPeriod) InTime(t time.Time) bool {
	nowweekday := int(t.Weekday())
	if t.Weekday() == time.Sunday {
		nowweekday = 7
	}
	for _, v := range p {
		if nowweekday == v {
			return true
		}
	}

	return false
}

//获取开启日
func (p WeekPeriod) GetStart() int {
	if len(p) > 0 {
		return p[0] //开启日
	}
	return -1
}

//获取结束日
func (p WeekPeriod) GetEnd() int {
	if len(p) > 0 {
		return p[len(p)-1]
	}
	return -1
}
