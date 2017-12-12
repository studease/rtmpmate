package Event

import (
	"fmt"
)

const (
	CANCEL    = "Event.cancel"
	CHANGE    = "Event.change"
	CLEAR     = "Event.clear"
	CLOSE     = "Event.close"
	COMPLETE  = "Event.complete"
	CONNECT   = "Event.connect"
	PUBLISH   = "Event.publish"
	UNPUBLISH = "Event.unpublish"
	RESIZE    = "Event.resize"
)

type Event struct {
	Type   string
	Target interface{} // A reference to the object that dispatched the event.
}

func New(Type string, Target interface{}) *Event {
	return &Event{Type, Target}
}

func (this *Event) Clone() *Event {
	return &Event{this.Type, this}
}

func (this *Event) ToString() string {
	return fmt.Sprintf("[Event type=%s]", this.Type)
}
