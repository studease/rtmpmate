package SignalEvent

import (
	"fmt"
	"studease.cn/events/Event"
)

const (
	SIGNAL = "SignalEvent.SIGNAL"
)

type SignalEvent struct {
	Event.Event
	Code int
}

func New(typ string, this interface{}, code int) *SignalEvent {
	return new(SignalEvent).Init(typ, this, code)
}

func (e *SignalEvent) Init(typ string, this interface{}, code int) *SignalEvent {
	e.Event.Init(typ, this)
	e.Code = code
	return e
}

func (e *SignalEvent) Clone() *SignalEvent {
	return New(e.Type, e.Target, e.Code)
}

func (e *SignalEvent) ToString() string {
	return fmt.Sprintf("[SignalEvent type=%s code=%d]", e.Type, e.Code)
}
