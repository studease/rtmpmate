package events

import (
	"container/list"
	"fmt"
	"reflect"
)

type listener struct {
	handler interface{}
	count   int
}

type EventDispatcher struct {
	listeners map[string]*list.List
}

func (this *EventDispatcher) AddEventListener(event string, handler interface{}, count int) {
	if this.listeners == nil {
		this.listeners = make(map[string]*list.List)
	}

	listeners, _ := this.listeners[event]
	if listeners == nil {
		listeners = list.New()
		this.listeners[event] = listeners
	}

	listeners.PushBack(&listener{handler, count})

	//fmt.Printf("Added event: %s, len=%d.\n", event, listeners.Len())
}

func (this *EventDispatcher) RemoveEventListener(event string, handler interface{}) {
	listeners, _ := this.listeners[event]
	if listeners == nil {
		return
	}

	if handler == nil {
		listeners.Init()
		return
	}

	for e := listeners.Front(); e != nil; e = e.Next() {
		ln := e.Value.(*listener)
		p1 := reflect.ValueOf(ln.handler).Pointer()
		p2 := reflect.ValueOf(handler).Pointer()
		if p1 == p2 {
			listeners.Remove(e)
			break
		}
	}

	//fmt.Printf("Removed event: %s, len: %d.\n", event, listeners.Len())
}

func (this *EventDispatcher) HasEventListener(event string) bool {
	listeners, _ := this.listeners[event]
	if listeners == nil || listeners.Len() == 0 {
		return false
	}

	return true
}

func (this *EventDispatcher) DispatchEvent(event interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Failed to DispatchEvent: %v.\n", err)
		}
	}()

	var eventValue = reflect.ValueOf(event)
	var eventElem = eventValue.Elem()
	var eventType = eventElem.FieldByName("Type").String()

	listeners, _ := this.listeners[eventType]
	if listeners == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Failed to handle %s: %v.\n", eventElem.MethodByName("ToString").Call(nil), err)
		}
	}()

	var eventIn = []reflect.Value{eventValue}

	for e := listeners.Front(); e != nil; e = e.Next() {
		ln := e.Value.(*listener)
		if ln.count > 0 {
			ln.count--

			if ln.count == 0 {
				listeners.Remove(e)
			}
		}

		if ln.handler != nil {
			reflect.ValueOf(ln.handler).Call(eventIn)
		}
	}
}
