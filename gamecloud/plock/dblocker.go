package plock

import (
	"time"

	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type DbLocker struct {
	url    string
	dbsess *mgo.Session
}

func NewDbLocker(dburl string) (*DbLocker, error) {
	var mongosess, err = mgo.Dial(dburl)
	if err != nil {
		return nil, err
	}

	return &DbLocker{dburl, mongosess}, nil
}

//重新拨号
func (l *DbLocker) _redial() error {
	var mongosess, err = mgo.Dial(l.url)
	if err != nil {
		return err
	}

	l.dbsess = mongosess
	return nil
}

func (l *DbLocker) Lock(sourcename string) {
	if err := l.dbsess.Ping(); err != nil { //socket abandon
		l.dbsess.Close()
		for {
			if err := l._redial(); err == nil {
				break
			} else {
				filelog.ERROR("plock", "plock redial fail : ", err.Error())
				time.Sleep(time.Second)
			}
		}
	}
	for {
		if err := l.dbsess.DB("plock").C("plock").Insert(bson.M{"_id": sourcename}); err != nil {
			//			filelog.INFO("plock", "Lock ", sourcename, "err : ", err.Error())
			time.Sleep(time.Millisecond * 10)
		} else {
			return
		}
	}
}

func (l *DbLocker) Unlock(sourcename string) {
	if err := l.dbsess.Ping(); err != nil { //socket abandon
		l.dbsess.Close()
		for {
			if err := l._redial(); err == nil {
				return
			} else {
				time.Sleep(time.Second)
			}
		}
	}
	for {
		if err := l.dbsess.DB("plock").C("plock").RemoveId(sourcename); err != nil {
			if err == mgo.ErrNotFound {
				filelog.ERROR("dblocker", "Unlock sourcename ", sourcename, " fail, remove source not found.")
				return
			} else {
				//				filelog.INFO("plock", "UnLock ", sourcename, "err : ", err.Error())
				time.Sleep(time.Millisecond * 10)
			}
		} else {
			return
		}
	}
}
