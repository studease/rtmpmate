package Event

import (
	"fmt"
)

const (
	CANCEL   = "Event.CANCEL"
	CHANGE   = "Event.CHANGE"
	CLEAR    = "Event.CLEAR"
	CLOSE    = "Event.CLOSE"
	COMPLETE = "Event.COMPLETE"
	CONNECT  = "Event.CONNECT"
	RESIZE   = "Event.RESIZE"
)

type Event struct {
	Type   string
	Target interface{}
}

func New(typ string, this interface{}) *Event {
	return new(Event).Init(typ, this)
}

func (e *Event) Init(typ string, this interface{}) *Event {
	e.Type = typ
	e.Target = this
	return e
}

func (e *Event) Clone() *Event {
	return New(e.Type, e.Target)
}

func (e *Event) ToString() string {
	return fmt.Sprintf("[Event type=%s]", e.Type)
}
