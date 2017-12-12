package NetStatusEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
)

const (
	NET_STATUS = "netStatus"
)

type NetStatusEvent struct {
	Event.Event
	Info map[string]interface{}
}

func New(Type string, Target interface{}, Info map[string]interface{}) *NetStatusEvent {
	return &NetStatusEvent{Event.Event{Type, Target}, Info}
}

func (this *NetStatusEvent) Clone() *NetStatusEvent {
	return &NetStatusEvent{Event.Event{this.Type, this}, this.Info}
}

func (this *NetStatusEvent) ToString() string {
	return fmt.Sprintf("[NetStatusEvent type=%s info=%v]", this.Type, this.Info)
}
