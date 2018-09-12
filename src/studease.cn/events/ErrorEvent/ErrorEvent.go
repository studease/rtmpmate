package ErrorEvent

import (
	"fmt"
	"studease.cn/events/Event"
)

const (
	ERROR = "ErrorEvent.ERROR"
)

type ErrorEvent struct {
	Event.Event
	Error error
}

func New(typ string, this interface{}, err error) *ErrorEvent {
	return new(ErrorEvent).Init(typ, this, err)
}

func (e *ErrorEvent) Init(typ string, this interface{}, err error) *ErrorEvent {
	e.Event.Init(typ, this)
	e.Error = err
	return e
}

func (e *ErrorEvent) Clone() *ErrorEvent {
	return New(e.Type, e.Target, e.Error)
}

func (e *ErrorEvent) ToString() string {
	return fmt.Sprintf("[ErrorEvent type=%s error=%v]", e.Type, e.Error)
}
