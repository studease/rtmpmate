package Chunk

import (
	"bytes"
	"rtmpmate.com/net/rtmp/Chunk/States"
)

type Chunk struct {
	BasicHeader
	MessageHeader
	Data bytes.Buffer

	CurrentFmt byte
	Polluted   bool
	Extended   bool
	Loaded     int
	State      byte
}

type BasicHeader struct {
	Fmt  byte   // 2 bits
	CSID uint32 // 6 | 14 | 22 bits
}

type MessageHeader struct {
	Timestamp       uint32 // 3 bytes
	MessageLength   int    // 3 bytes
	MessageTypeID   byte
	MessageStreamID uint32
}

func New() (*Chunk, error) {
	var c Chunk
	c.State = States.START

	return &c, nil
}
