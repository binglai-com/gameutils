package schedule

import (
	"errors"
	"time"
)

type job struct {
	desc   []uint32
	fun    func()
	lastup time.Time
}

//进程表 到点触发
type Schedule struct {
	list []*job
	_up  func(time.Time)
}

/*
增加时间任务
	timedesc []uint32 {时,分,秒,日,月,年}  //时分秒为必选项 ，日月年为选填项 且只可描述固定的缺省组合 例：
											日月年缺省 //描述为每天的  x时x分x秒
											月年缺省  //描述为每月的   x日x时x分x秒
											年缺省   //描述为每年的   x月x日x时x分x秒
											没有缺省项  //描述为一个指定的时间   x年x月x日 x时x分x秒
*/
func (s *Schedule) AddTask(timedesc []uint32, f func()) error {
	if len(timedesc) < 3 {
		return errors.New("Schedule.AddTask len(timedesc) < 3.")
	}

	s.list = append(s.list, &job{timedesc, f, time.Now()})
	return nil
}

//轮询任务
func (s *Schedule) Update(tt time.Time) {
	if s._up == nil {
		var lastsec int64 = 0
		s._up = func(t time.Time) {
			nowsec := t.Unix()
			if nowsec != lastsec { //没秒钟检查
				lastsec = nowsec
				t = time.Unix(nowsec, 0) //将纳秒设置为0进行比较
				for _, j := range s.list {
					var tar time.Time
					v := j.desc
					if len(v) == 3 { //周期为每天
						tar = time.Date(t.Year(), t.Month(), t.Day(), int(v[0]), int(v[1]), int(v[2]), 0, time.Local)
					} else if len(v) == 4 { //周期为每月
						tar = time.Date(t.Year(), t.Month(), int(v[3]), int(v[0]), int(v[1]), int(v[2]), 0, time.Local)
					} else if len(v) == 5 { //周期为每年
						tar = time.Date(t.Year(), time.Month(v[4]), int(v[3]), int(v[0]), int(v[1]), int(v[2]), 0, time.Local)
					} else if len(v) == 6 { //固定日期
						tar = time.Date(int(v[5]), time.Month(v[4]), int(v[3]), int(v[0]), int(v[1]), int(v[2]), 0, time.Local)
					} else {
						continue
					}

					if t.Equal(tar) || t.After(j.lastup) && t.Before(tar) {
						j.fun()
					}

					j.lastup = t
				}
			}
		}
	}

	s._up(tt)
}
