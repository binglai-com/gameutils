package gameutils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net/url"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/binglai-com/gameutils/proto"
	"github.com/globalsign/mgo/bson"
	"github.com/mohae/deepcopy"
)

//发送报警信息
func SendAlarm(desc string) {
	cmd := exec.Command("/bin/sh", "-c", "cagent_tools alarm '"+desc+"'")
	err := cmd.Run()
	if err != nil {
		filelog.ERROR("gameutils", "cmd.Run:", err.Error())
	}
}

//深度拷贝  被拷贝的结构体中不能出现循环引用的情况 如果有循环引用的结构体被深度拷贝会造成栈溢出  结构体中不能出现未导出的字段(已废弃)
func DeepCopy_Gob(dst, src interface{}) error {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

var deepcost = time.Duration(0)

//深度拷贝  被拷贝的结构体中不能出现循环引用的情况 如果有循环引用的结构体被深度拷贝会造成栈溢出 使用开源库 具体参考 https://github.com/mohae/deepcopy
func DeepCopy(src interface{}) interface{} {
	var start = time.Now()
	//	if res, err := bson.Marshal(src); err != nil {
	//		return err
	//	} else {
	//		mashalcost += time.Since(start)
	//		start = time.Now()
	//		if err := bson.Unmarshal(res, dst); err != nil {
	//			return err
	//		}
	//		unmarshalcost += time.Since(start)
	//	}
	var ret = deepcopy.Copy(src)
	deepcost += time.Since(start)
	return ret
}

type fileddesc struct {
	filedindex int
	keyname    string
}

//根据bson标记拷贝
func _copybybson(src interface{}, structdict map[reflect.Type][]fileddesc) interface{} {
	if src == nil {
		return nil
	}

	var rv = reflect.ValueOf(src)

	for {
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
			continue
		}
		break
	}

	if !rv.IsValid() {
		return nil
	}

	if rv.Kind() == reflect.Map {
		if rv.IsNil() {
			return nil
		}
		if rv.Type().Key().Kind() != reflect.String {
			return nil
		}
		var ret = make(bson.D, 0)
		for _, key := range rv.MapKeys() {
			ret = append(ret, bson.DocElem{key.String(), _copybybson(rv.MapIndex(key).Interface(), structdict)})
		}
		return ret
	} else if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.IsNil() {
			return nil
		}
		var ret = make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			ret[i] = _copybybson(rv.Index(i).Interface(), structdict)
		}
		return ret
	} else if rv.Kind() == reflect.Struct {
		var ret = make(bson.D, 0)
		var rt = rv.Type()
		var fields, ok = structdict[rt]
		if !ok {
			fields = make([]fileddesc, 0)
			for i := 0; i < rt.NumField(); i++ {
				var field = rt.Field(i)
				if field.PkgPath != "" { //该字段未导出
					continue
				}

				var tagname = field.Tag.Get("bson") //bson修饰为非导出字段
				if tagname == "-" {
					continue
				}

				if tagname == "" {
					tagname = strings.ToLower(field.Name)
				}

				fields = append(fields, fileddesc{i, tagname})
			}
			structdict[rt] = fields
		}

		for _, field := range fields {
			ret = append(ret, bson.DocElem{field.keyname, _copybybson(rv.Field(field.filedindex).Interface(), structdict)})
		}
		return ret
	} else if rv.Kind() == reflect.Func || rv.Kind() == reflect.Chan {
		return nil
	} else {
		return rv.Interface() //值类型  直接返回原值
	}

	return nil
}

//深度拷贝  被拷贝的结构体中不能出现循环引用的情况 如果有循环引用的结构体被深度拷贝会造成栈溢出 需要注意被拷贝的结构体中如有bson字段描述 比如 `bson:"-"` 那么该字段在深度拷贝时会被忽略
func DeepCopy2(dst, src interface{}) error {
	var start = time.Now()
	if res, err := bson.Marshal(src); err != nil {
		return err
	} else {
		if err := bson.Unmarshal(res, dst); err != nil {
			return err
		}
	}
	deepcost += time.Since(start)
	return nil
}
func GetDeepCopyCost() time.Duration {
	var ret1 = deepcost
	deepcost = time.Duration(0)
	return ret1
}

//根据bson标记获取字段值
func GetValueByBsonTag(src interface{}, bsontag string) interface{} {
	var rv = reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	var rt = rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		var ft = rt.Field(i)
		if ft.PkgPath != "" {
			continue
		}
		var tagname = ft.Tag.Get("bson") //bson修饰为非导出字段
		if tagname == "-" {
			continue
		}

		if tagname == "" {
			tagname = strings.ToLower(ft.Name)
		}
		if tagname == bsontag {
			return rv.Field(i).Interface()
		}
	}
	return nil
}

//根据结构体或结构体指针的字段名拷贝 拷贝出的结果的key是以结构体字段的bson标记命名的 比如  type t {PlayerName string `bson:"name"`} 那么根据'PlayerName' 拷贝出的结果就是 map["name"] = t.PlayerName
func CopyByFields(src interface{}, fields []string) (bson.D, error) {
	var rv = reflect.ValueOf(src)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("src expected type struct or struct ptr, got type %s ", reflect.TypeOf(src).Name())
	}

	var rt = rv.Type()
	var dictmap = make(map[reflect.Type][]fileddesc)
	if fields == nil {
		var ret = _copybybson(rv.Interface(), dictmap)
		if ret == nil {
			return bson.D{}, nil
		}

		return ret.(bson.D), nil
	} else {
		var ret = make(bson.D, 0)
		for _, fieldname := range fields {
			var ft, findfield = rt.FieldByName(fieldname)
			if !findfield {
				continue
			}

			//该字段未导出
			if ft.PkgPath != "" {
				continue
			}

			var tagname = ft.Tag.Get("bson") //bson修饰为非导出字段
			if tagname == "-" {
				continue
			}

			if tagname == "" {
				tagname = strings.ToLower(fieldname)
			}
			var fv = rv.FieldByName(fieldname)
			ret = append(ret, bson.DocElem{tagname, _copybybson(fv.Interface(), dictmap)})
		}
		return ret, nil
	}
}

type MsgEncoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	Simple_Encode(v interface{}) []byte
}



//binglai二进制序列化方法1
type BingLaiEncoder struct {
}

func (e *BingLaiEncoder) Encode(v interface{}) ([]byte, error) {
	return proto.Pack(v)
}

func (e *BingLaiEncoder) Decode(data []byte, v interface{}) error {
	return proto.UnPack(data, v)
}

func (e *BingLaiEncoder) Simple_Encode(v interface{}) []byte {
	s, _ := e.Encode(v)
	return s
}

//随机变量
func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min+1) + min
}

//是否是结构体
func IsStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

var sliceOfInts = reflect.TypeOf([]int(nil))
var sliceOfStrings = reflect.TypeOf([]string(nil))

// ParseForm will parse form values to struct via tag.
func ParseForm(form url.Values, obj interface{}) error {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	if !IsStructPtr(objT) {
		return fmt.Errorf("%v must be a struct pointer", obj)
	}
	objT = objT.Elem()
	objV = objV.Elem()
	for i := 0; i < objT.NumField(); i++ {
		fieldV := objV.Field(i)
		if !fieldV.CanSet() {
			continue
		}
		fieldT := objT.Field(i)
		tags := strings.Split(fieldT.Tag.Get("json"), ",")
		var tag string
		if len(tags) == 0 || len(tags[0]) == 0 {
			tag = fieldT.Name
		} else if tags[0] == "-" {
			continue
		} else {
			tag = tags[0]
		}

		value := form.Get(tag)
		if len(value) == 0 {
			continue
		}
		switch fieldT.Type.Kind() {
		case reflect.Bool:
			if strings.ToLower(value) == "on" || strings.ToLower(value) == "1" || strings.ToLower(value) == "yes" {
				fieldV.SetBool(true)
				continue
			}
			if strings.ToLower(value) == "off" || strings.ToLower(value) == "0" || strings.ToLower(value) == "no" {
				fieldV.SetBool(false)
				continue
			}
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			fieldV.SetBool(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetInt(x)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetUint(x)
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			fieldV.SetFloat(x)
		case reflect.Interface:
			fieldV.Set(reflect.ValueOf(value))
		case reflect.String:
			fieldV.SetString(value)
		case reflect.Struct:
			switch fieldT.Type.String() {
			case "time.Time":
				format := time.RFC3339
				if len(tags) > 1 {
					format = tags[1]
				}
				t, err := time.ParseInLocation(format, value, time.Local)
				if err != nil {
					return err
				}
				fieldV.Set(reflect.ValueOf(t))
			}
		case reflect.Slice:
			if fieldT.Type == sliceOfInts {
				formVals := form[tag]
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int(1))), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					val, err := strconv.Atoi(formVals[i])
					if err != nil {
						return err
					}
					fieldV.Index(i).SetInt(int64(val))
				}
			} else if fieldT.Type == sliceOfStrings {
				formVals := form[tag]
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf("")), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					fieldV.Index(i).SetString(formVals[i])
				}
			}
		}
	}
	return nil
}
