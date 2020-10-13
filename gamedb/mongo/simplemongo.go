package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func _simpledial(url string) (*mgo.Session, error) {
	if sess, err := mgo.Dial(url); err != nil {
		return nil, err
	} else {
		return sess, nil
	}
}

//查询一条记录
func FindOne(url string, dbname, colname string, find bson.M, res interface{}) error {
	sess, err := _simpledial(url)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Find(find).One(res)
}

//查询所有记录
func FindAll(url string, dbname, colname string, find bson.M, res interface{}) error {
	sess, err := _simpledial(url)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Find(find).All(res)
}

//查询所有记录
func FindAllLimit(url string, dbname, colname string, find bson.M, res interface{}, limit int) error {
	sess, err := _simpledial(url)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Find(find).Limit(limit).All(res)
}

//查询所有记录
func Count(url string, dbname, colname string, find bson.M) (int, error) {
	sess, err := _simpledial(url)
	if err != nil {
		return 0, err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Find(find).Count()
}

//插入一条记录
func Insert(url string, dbname, colname string, inserts interface{}) error {
	sess, err := _simpledial(url)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Insert(inserts)
}

//删除一条记录
func RemoveOne(url string, dbname, colname string, cond bson.M) error {
	sess, err := _simpledial(url)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).Remove(cond)
}

//删除一条记录
func RemoveAll(url string, dbname, colname string, cond bson.M) (*mgo.ChangeInfo, error) {
	sess, err := _simpledial(url)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	return sess.DB(dbname).C(colname).RemoveAll(cond)
}
