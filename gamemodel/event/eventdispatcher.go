package event

import (
	"sync"
)

var (
	_defautdispatcher *EventDispatcher = nil
)

func init() {
	_defautdispatcher = NewEventDispacher()
}

func Event(etype string, params ...interface{}) {
	_defautdispatcher.Event(etype, params...)
}

func OnSafe(etype string, h func(*EventHandler, ...interface{})) *EventHandler {
	return _defautdispatcher.OnSafe(etype, h)
}

func On(etype string, h func(*EventHandler, ...interface{})) *EventHandler {
	return _defautdispatcher.On(etype, h)
}

func Off(e *EventHandler) {
	_defautdispatcher.Off(e)
}

type EventDispatcher struct {
	_dispatching bool                        //是否正在派发中
	_events      map[string]*[]*EventHandler //线程不安全的事件 （需要使用者保证在单线程环境下注册和触发）
	sync.RWMutex
	_paralle_events map[string]*[]*EventHandler //线程安全，采用独立goroutine的方式触发的事件 可在多gorountine环境下使用

}

func NewEventDispacher() *EventDispatcher {
	return &EventDispatcher{
		false,
		make(map[string]*[]*EventHandler),
		sync.RWMutex{},
		make(map[string]*[]*EventHandler)}
}

//派发事件
func (this *EventDispatcher) Event(etype string, params ...interface{}) {
	this.RLock()
	pl, _ := this._paralle_events[etype]
	if pl != nil && len(*pl) > 0 {
		for _, e := range *pl {
			go e.handler(e, params...)
		}
	}
	this.RUnlock()

	l, _ := this._events[etype]
	if l != nil && len(*l) > 0 {
		this._dispatching = true
		var havedel = false
		for _, e := range *l {
			if !e.bodel {
				e.handler(e, params...)
			}
			if e.bodel {
				havedel = true
			}
		}

		if havedel { //清理所有已删除的事件
			for i := 0; i < len(*l); i++ {
				if (*l)[i].bodel {
					copy((*l)[i:], (*l)[i+1:])
					var n = len(*l) - 1
					(*l)[n] = nil
					*l = (*l)[:n]

					i--
				}
			}
		}
		this._dispatching = false //派发结束
	}

}

//监听事件 (线程安全)
func (this *EventDispatcher) OnSafe(etype string, h func(*EventHandler, ...interface{})) *EventHandler {
	this.Lock()
	defer this.Unlock()

	l, ok := this._paralle_events[etype]
	if !ok {
		l = &[]*EventHandler{}
		this._paralle_events[etype] = l
	}

	var e = &EventHandler{etype, h, true, false}
	*l = append(*l, e)
	return e
}

//监听事件 (非线程安全 使用者自己保证使用环境)
func (this *EventDispatcher) On(etype string, h func(*EventHandler, ...interface{})) *EventHandler {
	l, ok := this._events[etype]
	if !ok {
		l = &[]*EventHandler{}
		this._events[etype] = l
	}

	var e = &EventHandler{etype, h, false, false}
	*l = append(*l, e)
	return e
}

//注销事件
func (this *EventDispatcher) Off(e *EventHandler) {
	if e.boparalle {
		this.Lock()
		defer this.Unlock()
		l, ok := this._paralle_events[e._type]
		if ok {
			for i := 0; i < len(*l); i++ {
				if (*l)[i] == e {
					copy((*l)[i:], (*l)[i+1:])
					var n = len(*l) - 1
					(*l)[n] = nil
					*l = (*l)[:n]
					break
				}
			}
		}
	} else {
		if this._dispatching { //正在派发事件  只标记,在全部派发完以后再执行删除
			e.bodel = true
		} else { //非派发期间  直接删除
			l, ok := this._events[e._type]
			if ok {
				for i := 0; i < len(*l); i++ {
					if (*l)[i] == e {
						copy((*l)[i:], (*l)[i+1:])
						var n = len(*l) - 1
						(*l)[n] = nil
						*l = (*l)[:n]
						break
					}
				}
			}
		}
	}
}
