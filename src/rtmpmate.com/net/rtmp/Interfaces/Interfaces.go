package Interfaces

import (
	"io"
	"rtmpmate.com/net/rtmp/AMF"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/net/rtmp/Responder"
)

type IEventDispatcher interface {
	AddEventListener(event string, handler interface{}, count int)
	RemoveEventListener(event string, handler interface{})
	HasEventListener(event string) bool
	DispatchEvent(event interface{})
}

type INetConnection interface {
	io.ReadWriteCloser

	Connect(uri string, args ...*AMF.AMFValue) error
	CreateStream() error
	Call(method string, res *Responder.Responder, args ...*AMF.AMFValue) error
	WriteByChunk(b []byte, h *Message.Header) (int, error)

	GetAppName() string
	GetInstName() string
	GetFarID() string

	IEventDispatcher
}

type IStream interface {
	io.Closer

	Source(src IMuxer) error
	Sink(to IMuxer) error
	Play(name string, start float64, length float64, reset bool) error
	Record(mode string, maxDuration int, maxSize int) error
	Send(handler string, args ...*AMF.AMFValue) error
	GetDataFrame(name string) *DataMessage.DataMessage
	GetInitAudio() *AudioMessage.AudioMessage
	GetInitVideo() *VideoMessage.VideoMessage
	Clear() error
	Unlink(src IMuxer) error

	IEventDispatcher
}

type INetStream interface {
	io.Closer

	Attach(stream *IStream) error
	Play(name string) error
	Pause() error
	Resume() error
	ReceiveAudio(flag bool) error
	ReceiveVideo(flag bool) error
	Seek(offset float64) error
	Publish(name string, t string) error
	Send(handler string, args ...*AMF.AMFValue) error
	Dispose() error

	IEventDispatcher
}

type IMuxer interface {
	Source(src IStream) error
	Unlink(src IStream) error
	IsTypeSupported(mime string) bool
	EndOfStream(explain string)

	GetDataFrame(name string) *DataMessage.DataMessage
	GetInitAudio() *AudioMessage.AudioMessage
	GetInitVideo() *VideoMessage.VideoMessage

	IEventDispatcher
}
