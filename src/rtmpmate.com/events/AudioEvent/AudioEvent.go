package AudioEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
)

const (
	DATA = "AudioEvent.DATA"
)

type AudioEvent struct {
	Event.Event
	Message *AudioMessage.AudioMessage
}

func New(Type string, Target interface{}, m *AudioMessage.AudioMessage) *AudioEvent {
	return &AudioEvent{Event.Event{Type, Target}, m}
}

func (this *AudioEvent) Clone() *AudioEvent {
	return &AudioEvent{Event.Event{this.Type, this}, this.Message}
}

func (this *AudioEvent) ToString() string {
	return fmt.Sprintf("[AudioEvent type=%s data=%v]", this.Type, this.Message)
}
