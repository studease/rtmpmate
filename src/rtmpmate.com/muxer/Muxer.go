package muxer

import (
	"bytes"
	"fmt"
	"os"
	AACTypes "rtmpmate.com/codec/AAC/Types"
	"rtmpmate.com/codec/AudioFormats"
	H264Types "rtmpmate.com/codec/H264/Types"
	"rtmpmate.com/codec/VideoCodecs"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/NetStatusEvent"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/events/VideoEvent"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"syscall"
)

type Muxer struct {
	Dir                string
	Name               string
	Type               string
	DataFrames         map[string]*DataMessage.DataMessage
	InitAudio          *AudioMessage.AudioMessage
	InitVideo          *VideoMessage.VideoMessage
	Data               bytes.Buffer
	LastAudioTimestamp uint32
	LastVideoTimestamp uint32
	Src                Interfaces.IStream
	endOfStream        bool

	events.EventDispatcher
}

func New(dir string, name string) (*Muxer, error) {
	var m Muxer

	err := m.Init(dir, name, MuxerTypes.RTMP)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *Muxer) Init(dir string, name string, t string) error {
	this.Dir = dir + name + "/"
	this.Name = name
	this.Type = t
	this.DataFrames = make(map[string]*DataMessage.DataMessage)

	if _, err := os.Stat(this.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(this.Dir, os.ModeDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Muxer) Source(src Interfaces.IStream) error {
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

func (this *Muxer) Unlink(src Interfaces.IStream) error {
	src.RemoveEventListener(DataEvent.SET_DATA_FRAME, this.onSetDataFrame)
	src.RemoveEventListener(DataEvent.CLEAR_DATA_FRAME, this.onClearDataFrame)
	src.RemoveEventListener(AudioEvent.DATA, this.onAudio)
	src.RemoveEventListener(VideoEvent.DATA, this.onVideo)
	this.Src = nil

	return nil
}

func (this *Muxer) IsTypeSupported(mime string) bool {
	return true
}

func (this *Muxer) GetDataFrame(name string) *DataMessage.DataMessage {
	data, _ := this.DataFrames[name]
	return data
}

func (this *Muxer) GetInitAudio() *AudioMessage.AudioMessage {
	return this.InitAudio
}

func (this *Muxer) GetInitVideo() *VideoMessage.VideoMessage {
	return this.InitVideo
}

func (this *Muxer) onSetDataFrame(e *DataEvent.DataEvent) {
	fmt.Printf("Muxer.%s: %s\n", e.Message.Key, e.Message.Data.ToString(0))

	this.DataFrames[e.Message.Key] = e.Message
	this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, e.Message))
}

func (this *Muxer) onClearDataFrame(e *DataEvent.DataEvent) {
	delete(this.DataFrames, e.Message.Key)
	this.DispatchEvent(DataEvent.New(DataEvent.CLEAR_DATA_FRAME, this, e.Message))
}

func (this *Muxer) onAudio(e *AudioEvent.AudioEvent) {
	if this.InitAudio == nil {
		if e.Message.Format == AudioFormats.AAC && e.Message.DataType == AACTypes.SPECIFIC_CONFIG {
			this.InitAudio = e.Message
		}
	}

	this.LastAudioTimestamp = e.Message.Timestamp
	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *Muxer) onVideo(e *VideoEvent.VideoEvent) {
	if this.InitVideo == nil {
		if e.Message.Codec == VideoCodecs.AVC && e.Message.DataType == H264Types.SEQUENCE_HEADER {
			this.InitVideo = e.Message
		}
	}

	this.LastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *Muxer) EndOfStream(explain string) {
	this.endOfStream = true
	this.DispatchEvent(NetStatusEvent.New(NetStatusEvent.NET_STATUS, this, map[string]interface{}{
		"level":       Level.STATUS,
		"code":        Code.NETSTREAM_PLAY_STOP,
		"description": "play stop",
	}))
}

func (this *Muxer) Save(name string, data []byte) error {
	var (
		f   *os.File
		err error
	)

	if _, err = os.Stat(name); os.IsExist(err) {
		f, err = os.Open(name)
	} else {
		f, err = os.Create(name)
	}

	if err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
}
