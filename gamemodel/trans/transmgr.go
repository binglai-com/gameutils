package trans

import (
	"sync"
	"time"
)

//事务容器
type TransMgr struct {
	_loc   sync.Locker
	_trans TransQue //事务队列
}

//新建一个事务管理器
func NewTransMgr(loc sync.Locker) *TransMgr {
	return &TransMgr{_loc: loc}
}

//新增一条事务记录并入队
func (tm *TransMgr) NewTrans(TranTyp uint16, params ...interface{}) *Trans {
	var newtran = NewTrans(TranTyp, params...)
	tm.Push(newtran)
	return newtran
}

//事务入队
func (tm *TransMgr) Push(tran *Trans) {
	if tm._loc != nil {
		tm._loc.Lock()
		defer tm._loc.Unlock()
	}

	tm._trans.Push(tran)
}

//事务出队 传0将导出全部
func (tm *TransMgr) Pop(num int) (ret []*Trans) {
	defer func() {
		if len(ret) > 0 {
			var now = time.Now()
			for _, trans := range ret { //记录事务受理时间
				trans.Start(now)
			}
		}
	}()
	if tm._loc != nil {
		tm._loc.Lock()
		defer tm._loc.Unlock()
	}

	ret = tm._trans.Pop(num)
	return
}

//获取下一个待处理的事务
func (tm *TransMgr) Next() (ret *Trans) {
	defer func() {
		if ret != nil {
			ret.Start(time.Now())
		}
	}()
	if tm._loc != nil {
		tm._loc.Lock()
		defer tm._loc.Unlock()
	}
	ret = tm._trans.Next()
	return
}
