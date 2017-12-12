package UserControlEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage"
)

const (
	STREAM_BEGIN       = "StreamBegin"
	STREAM_EOF         = "StreamEOF"
	STREAM_DRY         = "StreamDry"
	SET_BUFFER_LENGTH  = "SetBufferLength"
	STREAM_IS_RECORDED = "StreamIsRecorded"
	PING_REQUEST       = "PingRequest"
	PING_RESPONSE      = "PingResponse"
)

type UserControlEvent struct {
	Event.Event
	Message *UserControlMessage.UserControlMessage
}

func New(Type string, Target interface{}, m *UserControlMessage.UserControlMessage) *UserControlEvent {
	return &UserControlEvent{Event.Event{Type, Target}, m}
}

func (this *UserControlEvent) Clone() *UserControlEvent {
	return &UserControlEvent{Event.Event{this.Type, this}, this.Message}
}

func (this *UserControlEvent) ToString() string {
	return fmt.Sprintf("[UserControlEvent type=%s message=%v]", this.Type, this.Message)
}
