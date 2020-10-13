package event

//事件句柄
type EventHandler struct {
	_type     string //事件类型
	handler   func(*EventHandler, ...interface{})
	boparalle bool //是否是平行事件 （true表示触发时会创建独立的goroutine进行调用 否则会在派发事件的协程内进行调用）
	bodel     bool //是否被删除了
}
