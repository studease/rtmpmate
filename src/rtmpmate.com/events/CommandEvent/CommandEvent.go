package CommandEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
)

const (
	CONNECT       = "connect"
	CLOSE         = "close"
	CREATE_STREAM = "createStream"
	RESULT        = "_result"
	ERROR         = "_error"

	PLAY          = "play"
	PLAY2         = "play2"
	DELETE_STREAM = "deleteStream"
	CLOSE_STREAM  = "closeStream"
	RECEIVE_AUDIO = "receiveAudio"
	RECEIVE_VIDEO = "receiveVideo"
	PUBLISH       = "publish"
	SEEK          = "seek"
	PAUSE         = "pause"
	ON_STATUS     = "onStatus"

	CHECK_BANDWIDTH = "checkBandwidth"
	GET_STATS       = "getStats"
)

type CommandEvent struct {
	Event.Event
	Message *CommandMessage.CommandMessage
}

func New(Type string, Target interface{}, m *CommandMessage.CommandMessage) *CommandEvent {
	return &CommandEvent{Event.Event{Type, Target}, m}
}

func (this *CommandEvent) Clone() *CommandEvent {
	return &CommandEvent{Event.Event{this.Type, this}, this.Message}
}

func (this *CommandEvent) ToString() string {
	return fmt.Sprintf("[CommandEvent type=%s message=%v]", this.Type, this.Message)
}
