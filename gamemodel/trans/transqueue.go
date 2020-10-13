package trans

import (
	"fmt"
)

type (
	TransQue   []*Trans
	TransGroup map[string]*TransQue
)

//事务入队
func (tq *TransQue) Push(add *Trans) {
	*tq = append(*tq, add)
}

//事务出队
func (tq *TransQue) Pop(num int) (ret []*Trans) {
	if len(*tq) == 0 {
		return
	}

	if num == 0 || num >= len(*tq) { //全部出队
		ret = append(ret, (*tq)...)
		*tq = (*tq)[:0]
	} else {
		ret = append(ret, (*tq)[:num]...)
		*tq = (*tq)[num:]
	}

	return
}

//下一个事务
func (tq *TransQue) Next() *Trans {
	if len(*tq) > 0 {
		return tq.Pop(1)[0]
	} else {
		return nil
	}
}

//根据条件分组
func (tq TransQue) Group(groupfun func(e *Trans) (groupkey string)) TransGroup {
	if len(tq) == 0 {
		return nil
	}
	var ret = make(TransGroup)
	for _, t := range tq {
		var gkey = groupfun(t)
		arr, find := ret[gkey]
		if !find {
			var tmp = make(TransQue, 0)
			ret[gkey] = &tmp
			arr = &tmp
		}
		*arr = append(*arr, t)
	}
	return ret
}

//根据groupkey获取事务队列
func (tg TransGroup) Find(groupkey string) TransQue {
	if que, find := tg[groupkey]; !find {
		return nil
	} else {
		return *que
	}
}

//根据事务类型分组
func _groupbytyp(e *Trans) string {
	return fmt.Sprintf("%d", e.Type)
}

//根据事务类型分组
func (tq TransQue) GroupByTyp() TransGroup {
	return tq.Group(_groupbytyp)
}

//根据事务类型获取事务队列
func (tg TransGroup) FindByTyp(trantyp uint16) TransQue {
	return tg.Find(fmt.Sprintf("%d", trantyp))
}
