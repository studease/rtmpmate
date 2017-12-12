package UserControlMessage

import (
	"encoding/binary"
	"fmt"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage/Event"
	EventTypes "rtmpmate.com/net/rtmp/Message/UserControlMessage/Event/Types"
)

type UserControlMessage struct {
	Message.Header
	Event Event.Event
}

func New() (*UserControlMessage, error) {
	var m UserControlMessage
	m.Type = Types.USER_CONTROL

	return &m, nil
}

func (this *UserControlMessage) Parse(b []byte, offset int, size int) error {
	if size < 10 {
		return fmt.Errorf("data (size=%d) not enough", size)
	}

	cost := 0

	this.Event.Type = binary.BigEndian.Uint16(b[offset+cost : offset+cost+2])
	cost += 2

	data := b[offset+cost:]

	switch this.Event.Type {
	case EventTypes.SET_BUFFER_LENGTH:
		this.Event.BufferLength = binary.BigEndian.Uint32(data[4:])
		cost += 4
		fallthrough
	case EventTypes.STREAM_BEGIN:
		fallthrough
	case EventTypes.STREAM_EOF:
		fallthrough
	case EventTypes.STREAM_DRY:
		fallthrough
	case EventTypes.STREAM_IS_RECORDED:
		this.Event.StreamID = binary.BigEndian.Uint32(data)
		cost += 4

	case EventTypes.PING_REQUEST:
		fallthrough
	case EventTypes.PING_RESPONSE:
		this.Event.Timestamp = binary.BigEndian.Uint32(data)
		cost += 4

	default:
	}

	return nil
}
