package FMP4Muxer

import (
	"fmt"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/muxer"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"rtmpmate.com/net/rtmp/Interfaces"
	"syscall"
)

type FMP4Muxer struct {
	muxer.Muxer
	MaxBufferLength int // bytes
	MaxBufferTime   int // ms
	LowLatency      bool
	Record          bool
}

func New(dir string, name string) (*FMP4Muxer, error) {
	var m FMP4Muxer

	err := m.Init(dir, name, MuxerTypes.FMP4)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *FMP4Muxer) Init(dir string, name string, t string) error {
	err := this.Muxer.Init(dir, name, t)
	if err != nil {
		return err
	}

	this.MaxBufferLength = 0x00200000
	this.MaxBufferTime = 3000
	this.LowLatency = true
	this.Record = true

	return nil
}

func (this *FMP4Muxer) Source(src Interfaces.IStream) error {
	if src == nil {
		return syscall.EINVAL
	}

	this.Src = src
	this.Src.AddEventListener(DataEvent.SET_DATA_FRAME, this.onSetDataFrame, 0)
	this.Src.AddEventListener(DataEvent.CLEAR_DATA_FRAME, this.onClearDataFrame, 0)
	this.Src.AddEventListener(AudioEvent.DATA, this.onAudio, 0)
	this.Src.AddEventListener(VideoEvent.DATA, this.onVideo, 0)

	m := this.Src.GetDataFrame("onMetaData")
	if m != nil {
		this.DataFrames["onMetaData"] = m
		this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, m))
	}

	return nil
}

func (this *FMP4Muxer) Unlink(src Interfaces.IStream) error {
	src.RemoveEventListener(DataEvent.SET_DATA_FRAME, this.onSetDataFrame)
	src.RemoveEventListener(DataEvent.CLEAR_DATA_FRAME, this.onClearDataFrame)
	src.RemoveEventListener(AudioEvent.DATA, this.onAudio)
	src.RemoveEventListener(VideoEvent.DATA, this.onVideo)
	this.Src = nil

	return nil
}

func (this *FMP4Muxer) onSetDataFrame(e *DataEvent.DataEvent) {
	fmt.Printf("FMP4Muxer.%s: %s\n", e.Message.Key, e.Message.Data.ToString(0))

	this.DataFrames[e.Message.Key] = e.Message
	this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, e.Message))
}

func (this *FMP4Muxer) onClearDataFrame(e *DataEvent.DataEvent) {
	delete(this.DataFrames, e.Message.Key)
	this.DispatchEvent(DataEvent.New(DataEvent.CLEAR_DATA_FRAME, this, e.Message))
}

func (this *FMP4Muxer) onAudio(e *AudioEvent.AudioEvent) {
	this.LastAudioTimestamp = e.Message.Timestamp
	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *FMP4Muxer) onVideo(e *VideoEvent.VideoEvent) {
	this.LastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *FMP4Muxer) EndOfStream(explain string) {
	this.Muxer.EndOfStream(explain)
}
