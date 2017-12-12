package VideoEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
)

const (
	DATA = "VideoEvent.DATA"
)

type VideoEvent struct {
	Event.Event
	Message *VideoMessage.VideoMessage
}

func New(Type string, Target interface{}, m *VideoMessage.VideoMessage) *VideoEvent {
	return &VideoEvent{Event.Event{Type, Target}, m}
}

func (this *VideoEvent) Clone() *VideoEvent {
	return &VideoEvent{Event.Event{this.Type, this}, this.Message}
}

func (this *VideoEvent) ToString() string {
	return fmt.Sprintf("[VideoEvent type=%s data=%v]", this.Type, this.Message)
}
