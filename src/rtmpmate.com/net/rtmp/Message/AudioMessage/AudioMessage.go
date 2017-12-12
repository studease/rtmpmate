package AudioMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AudioMessage struct {
	Message.Header
	Format     byte // 1111 0000
	SampleRate byte // 0000 1100
	SampleSize byte // 0000 0010
	Channels   byte // 0000 0001
	DataType   byte
	Payload    []byte
}

func New() (*AudioMessage, error) {
	var m AudioMessage
	m.Type = Types.AUDIO

	return &m, nil
}

func (this *AudioMessage) Parse(b []byte, offset int, size int) error {
	cost := 0

	this.Length = size

	tmp := b[offset+cost]
	this.Format = (tmp >> 4) & 0x0F
	this.SampleRate = (tmp >> 2) & 0x03
	this.SampleSize = (tmp >> 1) & 0x01
	this.Channels = tmp & 0x01
	cost++

	this.DataType = b[offset+cost]
	this.Payload = b[offset : offset+size]

	return nil
}
