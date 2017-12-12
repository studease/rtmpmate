package BandwidthMessage

import (
	"encoding/binary"
	"fmt"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type BandwidthMessage struct {
	Message.Header
	AckWindowSize uint32
	LimitType     byte
}

func New() (*BandwidthMessage, error) {
	var m BandwidthMessage
	m.Type = Types.BANDWIDTH

	return &m, nil
}

func (this *BandwidthMessage) Parse(b []byte, offset int, size int) error {
	if size < 5 {
		return fmt.Errorf("data (size=%d) not enough", size)
	}

	cost := 0

	this.AckWindowSize = binary.BigEndian.Uint32(b[offset+cost : offset+cost+4])
	cost += 4

	this.LimitType = b[offset+cost]
	cost += 1

	return nil
}
