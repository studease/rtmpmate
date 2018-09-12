package TimerEvent

import (
	"fmt"
	"studease.cn/events/Event"
)

const (
	TIMER          = "TimerEvent.TIMER"
	TIMER_COMPLETE = "TimerEvent.TIMER_COMPLETE"
)

type TimerEvent struct {
	Event.Event
}

func New(typ string, this interface{}) *TimerEvent {
	return new(TimerEvent).Init(typ, this)
}

func (e *TimerEvent) Init(typ string, this interface{}) *TimerEvent {
	e.Event.Init(typ, this)
	return e
}

func (e *TimerEvent) Clone() *TimerEvent {
	return New(e.Type, e.Target)
}

func (e *TimerEvent) ToString() string {
	return fmt.Sprintf("[TimerEvent type=%s]", e.Type)
}
