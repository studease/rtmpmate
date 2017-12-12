package Message

import (
	"encoding/binary"
	"fmt"
	"rtmpmate.com/net/rtmp/Chunk"
)

type Message struct {
	Header
	Payload []byte
}

type Header struct {
	Chunk.BasicHeader
	Type      byte
	Length    int // 3 bytes
	Timestamp uint32
	StreamID  uint32 // 3 bytes
}

func New() (*Message, error) {
	var m Message
	return &m, nil
}

func (this *Message) Parse(b []byte, offset int, size int) error {
	if size < 11 {
		return fmt.Errorf("data (size=%d) not enough", size)
	}

	cost := 0

	this.Type = b[offset+cost]
	cost += 1

	this.Length = int(b[offset+cost])<<16 | int(b[offset+cost+1])<<8 | int(b[offset+cost+2])
	cost += 3

	this.Timestamp = binary.BigEndian.Uint32(b[offset+cost : offset+cost+4])
	cost += 4

	this.StreamID = uint32(b[offset+cost])<<16 | uint32(b[offset+cost+1])<<8 | uint32(b[offset+cost+2])
	cost += 3

	if size-cost < this.Length {
		return fmt.Errorf("data (size=%d) not enough", size-cost)
	}

	this.Payload = b[offset+cost : offset+cost+this.Length]

	return nil
}
