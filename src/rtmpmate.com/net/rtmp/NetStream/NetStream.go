package NetStream

import (
	"fmt"
	"math"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/NetStatusEvent"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/events/UserControlEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/AMF"
	AMFTypes "rtmpmate.com/net/rtmp/AMF/Types"
	"rtmpmate.com/net/rtmp/Chunk/CSIDs"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/ObjectEncoding"
	"rtmpmate.com/net/rtmp/Stream"
	StreamTypes "rtmpmate.com/net/rtmp/Stream/Types"
)

type NetStream struct {
	Nc     *NetConnection.NetConnection
	Stream *Stream.Stream

	events.EventDispatcher
}

func New(nc *NetConnection.NetConnection) (*NetStream, error) {
	var ns NetStream
	ns.Nc = nc

	nc.AddEventListener(UserControlEvent.SET_BUFFER_LENGTH, ns.onSetBufferLength, 0)

	nc.AddEventListener(CommandEvent.CLOSE, ns.onClose, 0)
	nc.AddEventListener(CommandEvent.CREATE_STREAM, ns.onCreateStream, 0)
	nc.AddEventListener(CommandEvent.PLAY, ns.onPlay, 0)
	nc.AddEventListener(CommandEvent.PLAY2, ns.onPlay2, 0)
	nc.AddEventListener(CommandEvent.DELETE_STREAM, ns.onDeleteStream, 0)
	nc.AddEventListener(CommandEvent.CLOSE_STREAM, ns.onCloseStream, 0)
	nc.AddEventListener(CommandEvent.RECEIVE_AUDIO, ns.onReceiveAV, 0)
	nc.AddEventListener(CommandEvent.RECEIVE_VIDEO, ns.onReceiveAV, 0)
	nc.AddEventListener(CommandEvent.PUBLISH, ns.onPublish, 0)
	nc.AddEventListener(CommandEvent.SEEK, ns.onSeek, 0)
	nc.AddEventListener(CommandEvent.PAUSE, ns.onPause, 0)

	nc.AddEventListener(DataEvent.SET_DATA_FRAME, ns.onSetDataFrame, 0)
	nc.AddEventListener(DataEvent.CLEAR_DATA_FRAME, ns.onClearDataFrame, 0)
	nc.AddEventListener(AudioEvent.DATA, ns.onAudio, 0)
	nc.AddEventListener(VideoEvent.DATA, ns.onVideo, 0)

	return &ns, nil
}

func (this *NetStream) Attach(src *Stream.Stream) error {
	this.Stream = src
	return nil
}

func (this *NetStream) Play(name string) error {
	return nil
}

func (this *NetStream) Pause() error {
	return nil
}

func (this *NetStream) Resume() error {
	return nil
}

func (this *NetStream) ReceiveAudio(flag bool) error {
	return nil
}

func (this *NetStream) ReceiveVideo(flag bool) error {
	return nil
}

func (this *NetStream) Seek(offset float64) error {
	return nil
}

func (this *NetStream) Publish(name string, publishType string) error {
	return nil
}

func (this *NetStream) Send(handler string, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(handler)

	for _, v := range args {
		encoder.EncodeValue(v)
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	if this.Nc.ObjectEncoding == ObjectEncoding.AMF0 {
		h.Type = Types.DATA
	} else {
		h.Type = Types.AMF3_DATA
	}
	h.Fmt = 0
	h.CSID = CSIDs.COMMAND_2
	h.Length = encoder.Len()
	h.Timestamp = 0
	h.StreamID = uint32(this.Stream.ID)

	_, err = this.Nc.WriteByChunk(b, &h)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetStream) sendDataFrame(e *DataEvent.DataEvent) error {
	fmt.Printf("Sending %s...\n", e.Type)

	return this.Send(e.Message.Key, &AMF.AMFValue{
		Type: AMFTypes.ECMA_ARRAY,
		Data: e.Message.Data.Data,
	})
}

func (this *NetStream) clearDataFrame(e *DataEvent.DataEvent) error {
	return this.Send(e.Type, &AMF.AMFValue{
		Type: AMFTypes.STRING,
		Data: e.Message.Key,
	})
}

func (this *NetStream) sendAudio(e *AudioEvent.AudioEvent) error {
	//fmt.Printf("audio: %v\n", e.Message.Header)
	_, err := this.Nc.WriteByChunk(e.Message.Payload, &e.Message.Header)
	return err
}

func (this *NetStream) sendVideo(e *VideoEvent.VideoEvent) error {
	//fmt.Printf("video: %v\n", e.Message.Header)
	_, err := this.Nc.WriteByChunk(e.Message.Payload, &e.Message.Header)
	return err
}

func (this *NetStream) SendStatus(e *CommandEvent.CommandEvent, info *AMF.AMFObject) {
	var encoder AMF.Encoder
	encoder.EncodeString(Commands.ON_STATUS)
	encoder.EncodeNumber(0)
	encoder.EncodeNull()
	encoder.EncodeObject(info)

	e.Message.CSID = CSIDs.COMMAND_2
	this.Nc.SendEncodedBuffer(&encoder, e.Message.Header)
}

func (this *NetStream) Close() error {
	if this.Stream != nil {
		this.Stream.Close()
	}

	this.DispatchEvent(CommandEvent.New(CommandEvent.CLOSE, this, nil))

	return nil
}

func (this *NetStream) Dispose() error {
	if this.Stream != nil {
		this.Stream.Close()
		this.Stream.Clear()
	}

	return nil
}

func (this *NetStream) onSetBufferLength(e *UserControlEvent.UserControlEvent) {
	if this.Stream != nil {
		this.Stream.BufferTime = float64(e.Message.Event.BufferLength)
	}
}

func (this *NetStream) onCreateStream(e *CommandEvent.CommandEvent) {
	var command, code, description string
	if this.Nc.ReadAccess == "/" || this.Nc.ReadAccess == "/"+this.Nc.AppName {
		stream, _ := Stream.New(this.Nc.FarID)
		if stream != nil {
			stream.ID = 1 // ID 0 is used as NetConnection
			stream.Type = StreamTypes.IDLE
			this.Attach(stream)

			command = Commands.RESULT
		} else {
			command = Commands.ERROR
			code = Code.NETSTREAM_FAILED
			description = "Internal error"
		}
	} else {
		// TODO: Test on AMS
		command = Commands.ERROR
		code = Code.NETSTREAM_PLAY_FAILED
		description = "No read access"
	}

	var encoder AMF.Encoder
	encoder.EncodeString(command)
	encoder.EncodeNumber(math.Float64frombits(e.Message.TransactionID))
	encoder.EncodeNull()

	if command == Commands.RESULT {
		encoder.EncodeNumber(float64(this.Stream.ID))
	} else {
		// TODO: Test on AMS
		info, _ := this.Nc.GetInfoObject(Level.ERROR, code, description)
		encoder.EncodeObject(info)
	}

	this.Nc.SendEncodedBuffer(&encoder, e.Message.Header)
}

func (this *NetStream) onPlay(e *CommandEvent.CommandEvent) {
	this.Stream.Name = e.Message.StreamName
	this.Stream.AddEventListener(DataEvent.SET_DATA_FRAME, this.sendDataFrame, 0)
	this.Stream.AddEventListener(DataEvent.CLEAR_DATA_FRAME, this.clearDataFrame, 0)
	this.Stream.AddEventListener(AudioEvent.DATA, this.sendAudio, 0)
	this.Stream.AddEventListener(VideoEvent.DATA, this.sendVideo, 0)

	this.DispatchEvent(CommandEvent.New(CommandEvent.PLAY, this, e.Message))

	if this.Stream.Type == StreamTypes.IDLE {
		this.Stream.RemoveEventListener(DataEvent.SET_DATA_FRAME, this.sendDataFrame)
		this.Stream.RemoveEventListener(DataEvent.CLEAR_DATA_FRAME, this.clearDataFrame)
		this.Stream.RemoveEventListener(AudioEvent.DATA, this.sendAudio)
		this.Stream.RemoveEventListener(VideoEvent.DATA, this.sendVideo)
	}
}

func (this *NetStream) onPlay2(e *CommandEvent.CommandEvent) {

}

func (this *NetStream) onDeleteStream(e *CommandEvent.CommandEvent) {
	this.Stream.Close()
	this.Stream.Clear()
	this.Stream = nil
}

func (this *NetStream) onCloseStream(e *CommandEvent.CommandEvent) {
	this.Stream.Close()
	this.Stream.Clear()
	this.Stream = nil
}

func (this *NetStream) onReceiveAV(e *CommandEvent.CommandEvent) {
	if e.Message.Name == CommandEvent.RECEIVE_AUDIO && this.Stream.ReceiveAudio ||
		e.Message.Name == CommandEvent.RECEIVE_VIDEO && this.Stream.ReceiveVideo {
		return
	}

	if e.Message.Flag {
		info, _ := this.Nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_SEEK_NOTIFY, "Seek notify")
		this.SendStatus(e, info)

		info, _ = this.Nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_START, "Play start")
		this.SendStatus(e, info)
	}
}

func (this *NetStream) onPublish(e *CommandEvent.CommandEvent) {
	this.Stream.Name = e.Message.PublishingName
	this.Stream.RemoveEventListener(DataEvent.SET_DATA_FRAME, this.sendDataFrame)
	this.Stream.RemoveEventListener(DataEvent.CLEAR_DATA_FRAME, this.clearDataFrame)
	this.Stream.RemoveEventListener(AudioEvent.DATA, this.sendAudio)
	this.Stream.RemoveEventListener(VideoEvent.DATA, this.sendVideo)

	this.DispatchEvent(CommandEvent.New(CommandEvent.PUBLISH, this, e.Message))
}

func (this *NetStream) onSeek(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if this.Stream.Type == StreamTypes.PLAYING_VOD {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.Stream.Duration {
			info, _ = this.Nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_SEEK_NOTIFY, "Seek notify")
		} else {
			info, _ = this.Nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_SEEK_INVALIDTIME, "Seek invalid time")
		}
	} else {
		info, _ = this.Nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_SEEK_FAILED, "Seek failed")
	}

	this.SendStatus(e, info)
}

func (this *NetStream) onPause(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if e.Message.Pause {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.Stream.Duration {
			this.Stream.Pause = e.Message.Pause
			this.Stream.CurrentTime = e.Message.MilliSeconds

			info, _ = this.Nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PAUSE_NOTIFY, "Pause notify")
		} else {
			info, _ = this.Nc.GetInfoObject(Level.ERROR, "Pause invalid time", "Pause invalid time")
		}
	} else {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.Stream.Duration {
			this.Stream.Pause = e.Message.Pause
			this.Stream.CurrentTime = e.Message.MilliSeconds

			info, _ = this.Nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_UNPAUSE_NOTIFY, "Unpause notify")
		} else {
			info, _ = this.Nc.GetInfoObject(Level.ERROR, "Unpause invalid time", "Unpause invalid time")
		}
	}

	this.SendStatus(e, info)
}

func (this *NetStream) onStatus(e *NetStatusEvent.NetStatusEvent) {

}

func (this *NetStream) onSetDataFrame(e *DataEvent.DataEvent) {
	this.Stream.DispatchEvent(DataEvent.New(e.Type, this, e.Message))
}

func (this *NetStream) onClearDataFrame(e *DataEvent.DataEvent) {
	this.Stream.DispatchEvent(DataEvent.New(e.Type, this, e.Message))
}

func (this *NetStream) onAudio(e *AudioEvent.AudioEvent) {
	this.Stream.DispatchEvent(AudioEvent.New(e.Type, this, e.Message))
}

func (this *NetStream) onVideo(e *VideoEvent.VideoEvent) {
	this.Stream.DispatchEvent(VideoEvent.New(e.Type, this, e.Message))
}

func (this *NetStream) onMetaData(e *DataEvent.DataEvent) {
	//fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))
}

func (this *NetStream) onClose(e *CommandEvent.CommandEvent) {
	this.Close()
}
