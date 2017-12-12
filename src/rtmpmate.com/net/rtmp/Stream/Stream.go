package Stream

import (
	"container/list"
	AACTypes "rtmpmate.com/codec/AAC/Types"
	"rtmpmate.com/codec/AudioFormats"
	H264Types "rtmpmate.com/codec/H264/Types"
	"rtmpmate.com/codec/VideoCodecs"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/AMF"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage/FrameTypes"
	"rtmpmate.com/net/rtmp/Stream/RecordModes"
	"syscall"
)

type Stream struct {
	ID                 int
	Name               string
	Type               string
	Chunks             list.List
	DataFrames         map[string]*DataMessage.DataMessage
	InitAudio          *AudioMessage.AudioMessage
	InitVideo          *VideoMessage.VideoMessage
	Time               float64 // at when to init
	BufferTime         float64
	CurrentTime        float64
	Duration           float64
	MaxQueueDelay      float64 // ms
	MaxQueueSize       int
	Pause              bool
	PublishQueryString string
	ReceiveAudio       bool
	ReceiveVideo       bool

	src Interfaces.IMuxer
	to  Interfaces.IMuxer
	events.EventDispatcher
}

func New(name string) (*Stream, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	var s Stream
	s.Name = name
	s.DataFrames = make(map[string]*DataMessage.DataMessage)

	return &s, nil
}

func (this *Stream) Source(src Interfaces.IMuxer) error {
	if src == nil {
		return syscall.EINVAL
	}

	this.src = src
	this.src.AddEventListener(DataEvent.SET_DATA_FRAME, this.onSetDataFrame, 0)
	this.src.AddEventListener(DataEvent.CLEAR_DATA_FRAME, this.onClearDataFrame, 0)
	this.src.AddEventListener(AudioEvent.DATA, this.onAudio, 0)
	this.src.AddEventListener(VideoEvent.DATA, this.onVideo, 0)

	m := this.src.GetDataFrame("onMetaData")
	if m != nil {
		this.DataFrames["onMetaData"] = m
		this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, m))
	}

	return nil
}

func (this *Stream) Sink(to Interfaces.IMuxer) error {
	if to == nil {
		return syscall.EINVAL
	}

	this.to = to
	this.to.Source(this)

	return nil
}

func (this *Stream) Play(name string, start float64, length float64, reset bool) error {
	return nil
}

func (this *Stream) Record(mode string, maxDuration int, maxSize int) error {
	switch mode {
	case RecordModes.RECORD:
	case RecordModes.APPEND:
	case RecordModes.STOP:
	default:
		return syscall.EINVAL
	}

	return nil
}

func (this *Stream) Send(handler string, args ...*AMF.AMFValue) error {
	return nil
}

func (this *Stream) GetDataFrame(name string) *DataMessage.DataMessage {
	data, _ := this.DataFrames[name]
	return data
}

func (this *Stream) GetInitAudio() *AudioMessage.AudioMessage {
	return this.InitAudio
}

func (this *Stream) GetInitVideo() *VideoMessage.VideoMessage {
	return this.InitVideo
}

func (this *Stream) Clear() error {
	this.Chunks.Init()
	this.DataFrames = make(map[string]*DataMessage.DataMessage)
	this.InitAudio = nil
	this.InitVideo = nil

	return nil
}

func (this *Stream) Close() error {
	if this.src != nil {
		this.Unlink(this.src)
	}
	if this.to != nil {
		this.to.Unlink(this)
	}

	return nil
}

func (this *Stream) Unlink(src Interfaces.IMuxer) error {
	src.RemoveEventListener(DataEvent.SET_DATA_FRAME, this.onSetDataFrame)
	src.RemoveEventListener(DataEvent.CLEAR_DATA_FRAME, this.onClearDataFrame)
	src.RemoveEventListener(AudioEvent.DATA, this.onAudio)
	src.RemoveEventListener(VideoEvent.DATA, this.onVideo)
	this.src = nil

	return nil
}

func (this *Stream) onSetDataFrame(e *DataEvent.DataEvent) {
	this.DataFrames[e.Message.Key] = e.Message
	this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, e.Message))
}

func (this *Stream) onClearDataFrame(e *DataEvent.DataEvent) {
	delete(this.DataFrames, e.Message.Key)
	this.DispatchEvent(DataEvent.New(DataEvent.CLEAR_DATA_FRAME, this, e.Message))
}

func (this *Stream) onAudio(e *AudioEvent.AudioEvent) {
	if this.InitAudio == nil {
		if e.Message.Format == AudioFormats.AAC && e.Message.DataType == AACTypes.SPECIFIC_CONFIG {
			this.InitAudio = e.Message
			if this.InitVideo != nil {
				this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, this.InitVideo))
				this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, this.InitAudio))
			}
		}
	} else if this.InitVideo != nil {
		this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
	}
}

func (this *Stream) onVideo(e *VideoEvent.VideoEvent) {
	if this.InitVideo == nil {
		if e.Message.Codec == VideoCodecs.AVC && e.Message.DataType == H264Types.SEQUENCE_HEADER {
			this.InitVideo = e.Message
			if this.InitAudio != nil {
				this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, this.InitAudio))
				this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, this.InitVideo))
			}
		} else if e.Message.FrameType == FrameTypes.KEYFRAME || e.Message.FrameType == FrameTypes.GENERATED_KEYFRAME {
			if this.InitAudio == nil {
				this.InitAudio = this.src.GetInitAudio()
			}
			if this.InitVideo == nil {
				this.InitVideo = this.src.GetInitVideo()
			}

			if this.InitAudio != nil && this.InitVideo != nil {
				this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, this.InitAudio))
				this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, this.InitVideo))
			}

			this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
		}
	} else if this.InitAudio != nil {
		this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
	}
}
