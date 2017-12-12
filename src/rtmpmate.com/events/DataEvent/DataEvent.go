package DataEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
)

const (
	SET_DATA_FRAME   = "@setDataFrame"
	CLEAR_DATA_FRAME = "@clearDataFrame"
)

type DataEvent struct {
	Event.Event
	Message *DataMessage.DataMessage
}

func New(Type string, Target interface{}, m *DataMessage.DataMessage) *DataEvent {
	return &DataEvent{Event.Event{Type, Target}, m}
}

func (this *DataEvent) Clone() *DataEvent {
	return &DataEvent{Event.Event{this.Type, this}, this.Message}
}

func (this *DataEvent) ToString() string {
	return fmt.Sprintf("[DataEvent type=%s message=%v]", this.Type, this.Message)
}
