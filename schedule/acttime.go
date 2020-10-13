package schedule

import (
	"time"
)

const (
	ROUNDTYPE_FULL_DAY    uint8 = 1 + iota //每满整24小时 累计一天
	ROUNDTYPE_NATRUAL_DAY                  //每过一个自然天（每天0点） 累计一天
)

//获取当天0点
func GetZeroTime(t int64) time.Time {
	tmp := time.Unix(t, 0)
	return time.Date(tmp.Year(), tmp.Month(), tmp.Day(), 0, 0, 0, 0, time.Local)
}

//获取当前是活动第几天
func GetStartDay(starttime int64, nowtime int64, roundtype uint8) int {
	if starttime == 0 {
		return 0
	}

	if roundtype == ROUNDTYPE_FULL_DAY { //满24小时为一天
		return int(time.Since(time.Unix(starttime, 0))/(time.Hour*24) + 1)
	} else if roundtype == ROUNDTYPE_NATRUAL_DAY { //一个自然天为一天
		startzero := GetZeroTime(starttime)
		nowzero := GetZeroTime(nowtime)
		return int(nowzero.Sub(startzero)/(time.Hour*24) + 1)
	} else {
		return 0
	}
}

//获取第n天开始的时间
func CountDayStart(starttime int64, day int, roundtype uint8) time.Time {
	if roundtype == ROUNDTYPE_FULL_DAY {
		return time.Unix(starttime+int64(day-1)*86400, 0)
	} else if roundtype == ROUNDTYPE_NATRUAL_DAY { //一个自然天为一天
		startzero := GetZeroTime(starttime)
		return startzero.Add(time.Duration(day-1) * (time.Hour * 24))
	} else {
		return time.Time{}
	}
}
