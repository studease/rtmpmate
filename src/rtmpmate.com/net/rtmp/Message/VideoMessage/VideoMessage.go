package VideoMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type VideoMessage struct {
	Message.Header
	FrameType byte // 0xF0
	Codec     byte // 0x0F
	DataType  byte
	Payload   []byte
}

func New() (*VideoMessage, error) {
	var m VideoMessage
	m.Type = Types.VIDEO

	return &m, nil
}

func (this *VideoMessage) Parse(b []byte, offset int, size int) error {
	cost := 0

	this.Length = size

	tmp := b[offset+cost]
	this.FrameType = (tmp >> 4) & 0x0F
	this.Codec = tmp & 0x0F
	cost++

	this.DataType = b[offset+cost]
	this.Payload = b[offset : offset+size]

	return nil
}
