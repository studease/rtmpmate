package Meta

import ()

type Meta struct {
	Type      string
	ID        uint32
	Timescale uint32
	Duration  uint32

	CodecWidth    int
	CodecHeight   int
	PresentWidth  int
	PresentHeight int

	AVCC []byte

	ChannelCount  byte
	SampleRate    int
	ChannelConfig []byte
}
