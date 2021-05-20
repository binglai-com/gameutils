package mongo

import (
	"log"
	"testing"

	"github.com/globalsign/mgo/bson"
)

type Subject struct {
	Sub   string `bson:"sub"`
	Score int    `bson:"score"`
}

func TestUpdate(t *testing.T) {
	db := new(DataBase)
	if !db.Init("mongodb://127.0.0.1:27017/") {
		t.Error("db init fail")
	}

	db.DropCol("testmongo", "testupdate")

	db.Insert("testmongo", "testupdate", bson.M{"_id": 1, "name": "test1", "subjects": []Subject{{"math", 90}, {"ch", 95}, {"en", 80}}})
	db.Insert("testmongo", "testupdate", bson.M{"_id": 2, "name": "test2", "subjects": []Subject{{"math", 100}, {"ch", 80}, {"en", 99}}})

	err := db.Update("testmongo", "testupdate", []UpdateOps{
		{Query: bson.M{"_id": 2}, Update: bson.M{"$set": bson.M{"subjects.$[elem].score": 60}}, ArrayFilters: []bson.M{bson.M{"elem.sub": "ch"}}},
	})
	if err != nil {
		t.Error("test update fail : ", err.Error())
	}
}

func TestFindAndModify(t *testing.T) {
	db := new(DataBase)
	if !db.Init("mongodb://127.0.0.1:27017/") {
		t.Error("db init fail")
	}

	db.DropCol("testmongo", "findandmodify")

	//写入两条测试用的数据
	db.Insert("testmongo", "findandmodify", bson.M{"_id": 1, "name": "test1", "subjects": []Subject{{"math", 90}, {"ch", 95}, {"en", 80}}})
	db.Insert("testmongo", "findandmodify", bson.M{"_id": 2, "name": "test2", "subjects": []Subject{{"math", 100}, {"ch", 80}, {"en", 99}}})

	var res = bson.M{}
	err := db.FindAndModify("testmongo", "findandmodify", bson.M{"_id": 1}, FindAndModifyDecorater{
		Update: bson.M{"$set": bson.M{ /*"subjects.$[elem].score": 92,*/ "name": "test11"}, "$pull": bson.M{"subjects": bson.M{"sub": "ch"}}},
		// ArrayFilters: []bson.M{bson.M{"elem.sub": "math"}},
		New:    true,
		Fields: bson.M{"subjects": bson.M{"$elemMatch": bson.M{"sub": "math"}}},
	}, &res)
	log.Println(res)
	if err != nil {
		t.Error("Find And Modify err : ", err.Error())
	}
}
