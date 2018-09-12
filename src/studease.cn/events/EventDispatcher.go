package events

import (
	"container/list"
	"reflect"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
)

type eventListener struct {
	handler interface{}
	count   int
}

type EventDispatcher struct {
	listeners map[string]*list.List
	globals   *list.List
}

func (this *EventDispatcher) AddEventListener(event string, handler interface{}, count int) {
	if this.listeners == nil {
		this.listeners = make(map[string]*list.List)
	}

	l, _ := this.listeners[event]

	if l == nil {
		l = list.New()
		this.listeners[event] = l
	}

	l.PushBack(&eventListener{handler, count})

	log.Debug(level.EVENT, "add event: type=%s, len=%d", event, l.Len())
}

func (this *EventDispatcher) RemoveEventListener(event string, handler interface{}) {
	l, _ := this.listeners[event]
	if l == nil {
		return
	}

	if handler == nil {
		l.Init()
		return
	}

	h0 := reflect.ValueOf(handler).Pointer()

	for e := l.Front(); e != nil; e = e.Next() {
		ln := e.Value.(*eventListener)
		h1 := reflect.ValueOf(ln.handler).Pointer()

		if h0 == h1 {
			l.Remove(e)
			break
		}
	}

	log.Debug(level.EVENT, "remove event: type=%s, len=%d", event, l.Len())
}

func (this *EventDispatcher) AddGlobalListener(handler interface{}, count int) {
	if this.globals == nil {
		this.globals = list.New()
	}

	this.globals.PushBack(&eventListener{handler, count})

	log.Debug(level.EVENT, "add global event: len=%d", this.globals.Len())
}

func (this *EventDispatcher) RemoveGlobalListener(handler interface{}) {
	h0 := reflect.ValueOf(handler).Pointer()

	for e := this.globals.Front(); e != nil; e = e.Next() {
		ln := e.Value.(*eventListener)
		h1 := reflect.ValueOf(ln.handler).Pointer()

		if h0 == h1 {
			this.globals.Remove(e)
			break
		}
	}
}

func (this *EventDispatcher) HasEventListener(event string) bool {
	l, _ := this.listeners[event]
	return l != nil && l.Len() != 0
}

func (this *EventDispatcher) HasGlobalListener() bool {
	l := this.globals
	return l != nil && l.Len() != 0
}

func (this *EventDispatcher) DispatchEvent(event interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Debug(level.EVENT, "failed to DispatchEvent: %v", err)
		}
	}()

	val := reflect.ValueOf(event)
	ele := val.Elem()
	evt := ele.FieldByName("Type").String()

	l, _ := this.listeners[evt]

	if l == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Debug(level.EVENT, "failed to handle event %s: %v", ele.MethodByName("ToString").Call(nil), err)
		}
	}()

	this.dispatchEvent(l, val)
	this.dispatchEvent(this.globals, val)
}

func (this *EventDispatcher) dispatchEvent(l *list.List, v reflect.Value) {
	if l == nil {
		return
	}

	for e := l.Front(); e != nil; e = e.Next() {
		ln := e.Value.(*eventListener)

		if ln.count > 0 {
			ln.count--

			if ln.count == 0 {
				l.Remove(e)
			}
		}

		if ln.handler != nil {
			reflect.ValueOf(ln.handler).Call([]reflect.Value{v})
		}
	}
}
