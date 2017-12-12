package Track

import (
	"bytes"
	"container/list"
)

type Track struct {
	ID             uint32
	Samples        list.List
	SequenceNumber uint32
	Length         int
}

type Sample struct {
	Flags      Flags
	Units      list.List
	Length     int
	IsKeyFrame bool
	Duration   int
	Size       int
	CTS        int
	DTS        int
	PTS        int
}

type Flags struct {
	IsLeading     byte
	DependsOn     byte
	IsDependedOn  byte
	HasRedundancy byte
	IsNonSync     byte
}

type Unit struct {
	Type byte
	Data bytes.Buffer
}
