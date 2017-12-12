package SharedObjectMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/AMF"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/ObjectEncoding"
)

// Not sure about this.
type SharedObjectMessage struct {
	Message.Header
	Name    AMF.AMFString
	Version byte
	Flags   uint8
	Events  list.List
}

func New(encoding byte) (*SharedObjectMessage, error) {
	var m SharedObjectMessage

	if encoding == ObjectEncoding.AMF0 {
		m.Type = Types.DATA
	} else {
		m.Type = Types.AMF3_DATA
	}

	return &m, nil
}
