package ErrorEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
)

const (
	ERROR = "error"
)

type ErrorEvent struct {
	Event.Event
	Error error
}

func New(Type string, Target interface{}, err error) *ErrorEvent {
	return &ErrorEvent{Event.Event{Type, Target}, err}
}

func (this *ErrorEvent) Clone() *ErrorEvent {
	return &ErrorEvent{Event.Event{this.Type, this}, this.Error}
}

func (this *ErrorEvent) ToString() string {
	return fmt.Sprintf("[ErrorEvent type=%s error=%v]", this.Type, this.Error)
}
