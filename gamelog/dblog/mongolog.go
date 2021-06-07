package dblog

import (
	"fmt"
	"time"

	"github.com/binglai-com/gameutils/gamedb/mongo"
	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/globalsign/mgo/bson"
)

type Logger struct {
	db *mongo.DataBase
}

//初始化一个Logger
func NewLogger(mongourl string) (*Logger, error) {
	var ret = &Logger{}
	ret.db = new(mongo.DataBase)
	if !ret.db.Init(mongourl) {
		return nil, fmt.Errorf("dail mongo url %s fail.", mongourl)
	}

	return ret, nil
}

//记录日志
func (l *Logger) Log(dbname string, colname string, log bson.M) {
	log["time"] = time.Now().Unix()
	if err := l.db.Insert(dbname, colname, log); err != nil {
		filelog.ERROR("dblog", "log into ", dbname, ".", colname, " with ", log, " err : ", err.Error())
		return
	}
}
