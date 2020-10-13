package schedule

import (
	"time"
)

//游戏中的时间段
type GamePeriod [][]uint32

//判断t是否在时间段内
func (p GamePeriod) InTime(t time.Time) bool {
	if len(p) < 2 {
		return false
	}

	if len(p[0]) != len(p[1]) {
		return false
	}

	if len(p[0]) < 3 {
		return false
	}

	start, end := p.GetStart(), p.GetEnd()
	if t.After(start) && t.Before(end) {
		return true
	} else {
		return false
	}

}

//获取开启时间
func (p GamePeriod) GetStart() time.Time {
	var h, m, s = 0, 0, 0
	now := time.Now()
	d, mon, y := 0, 0, 0
	if len(p) > 0 {
		d, mon, y = now.Day(), int(now.Month()), now.Year()
		if len(p[0]) >= 3 {
			h, m, s = int(p[0][0]), int(p[0][1]), int(p[0][2])
		}
		if len(p[0]) > 3 {
			d = int(p[0][3])
		}
		if len(p[0]) > 4 {
			mon = int(p[0][4])
		}
		if len(p[0]) > 5 {
			y = int(p[0][5])
		}
	}
	return time.Date(y, time.Month(mon), d, h, m, s, 0, time.Local)
}

//获取结束时间
func (p GamePeriod) GetEnd() time.Time {
	var h, m, s = 0, 0, 0
	now := time.Now()
	d, mon, y := 0, 0, 0
	if len(p) > 1 {
		d, mon, y = now.Day(), int(now.Month()), now.Year()
		if len(p[1]) >= 3 {
			h, m, s = int(p[1][0]), int(p[1][1]), int(p[1][2])
		}
		if len(p[1]) > 3 {
			d = int(p[1][3])
		}
		if len(p[1]) > 4 {
			mon = int(p[1][4])
		}
		if len(p[1]) > 5 {
			y = int(p[1][5])
		}
	}
	return time.Date(y, time.Month(mon), d, h, m, s, 0, time.Local)
}
