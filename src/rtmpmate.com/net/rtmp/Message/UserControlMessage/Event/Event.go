package Event

import ()

type Event struct {
	Type         uint16
	StreamID     uint32
	BufferLength uint32
	Timestamp    uint32
}
