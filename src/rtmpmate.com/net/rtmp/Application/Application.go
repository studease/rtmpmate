package Application

import (
	"fmt"
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"rtmpmate.com/net/rtmp"
	"rtmpmate.com/net/rtmp/AMF"
	AMFTypes "rtmpmate.com/net/rtmp/AMF/Types"
	"rtmpmate.com/net/rtmp/Application/Instance"
	"rtmpmate.com/net/rtmp/Message/BandwidthMessage/LimitTypes"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	EventTypes "rtmpmate.com/net/rtmp/Message/UserControlMessage/Event/Types"
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/NetStream"
	StreamTypes "rtmpmate.com/net/rtmp/Stream/Types"
	"sync"
	"syscall"
)

var (
	apps    map[string]*Application
	appsMtx sync.RWMutex
)

func init() {
	apps = make(map[string]*Application)
}

type Application struct {
	Name string

	instances    map[string]*Instance.Instance
	instancesMtx sync.RWMutex

	statsToAdmin
	events.EventDispatcher
}

type stats struct {
	bytesIn  int
	bytesOut int

	msgIn      int
	msgOut     int
	msgDropped int
}

type statsToAdmin struct {
	stats

	launchTime uint64 // time the application started.
	upTime     uint64 // time, in seconds, the application has been running.

	// Total number of instances that have been loaded since the application started.
	// This property does not represent the total number of active instances loaded. To get the number of active
	// instances loaded, subtract the value of total_instances_unloaded from total_instances_loaded.
	totalInstancesLoaded   int
	totalInstancesUnloaded int // Total number of instances that have been unloaded since the application started.

	totalConnects    int // Total number of socket connections to the application since the application was started.
	totalDisconnects int // Total number of disconnections from the application since the application was started.

	accepted        int // Total number of connection attempts accepted by this application.
	rejected        int // Total number of connection attempts rejected by this application.
	connected       int // Total number of connections currently active.
	normalConnects  int // Total number of normal connections.
	virtualConnects int // Total number of connections through a remote edge.
	adminConnects   int // Total number of administrator connections.
	debugConnects   int // Total number of debug connections.

	swfVerificationAttempts int
	swfVerificationMatches  int
	swfVerificationFailures int
}

func Get(name string) (*Application, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	appsMtx.Lock()

	app, ok := apps[name]
	if ok {
		appsMtx.Unlock()
		return app, nil
	}

	var a Application
	a.Name = name
	a.instances = make(map[string]*Instance.Instance)

	app = &a
	apps[name] = app
	appsMtx.Unlock()

	app.onStart()

	return app, nil
}

func (this *Application) GetInstance(name string) (*Instance.Instance, error) {
	if name == "" {
		name = "_definst_"
	}

	this.instancesMtx.Lock()

	inst, ok := this.instances[name]
	if ok == false {
		inst, _ = Instance.New(this.Name, name)
		this.instances[name] = inst
	}

	this.instancesMtx.Unlock()

	return inst, nil
}

func (this *Application) Shutdown() {
	this.instancesMtx.Lock()

	for _, inst := range this.instances {
		inst.Unload()
		delete(this.instances, inst.Name)
	}

	this.instancesMtx.Unlock()
}

func (this *Application) GetStats() *stats {
	return &this.stats
}

func (this *Application) onStart() {

}

func (this *Application) onStop() {

}

func HandshakeComplete(conn net.Conn) (*NetConnection.NetConnection, error) {
	nc, err := NetConnection.New(conn)
	if err != nil {
		fmt.Printf("Failed to create NetConnection: %v.\n", err)
		return nil, err
	}

	nc.AddEventListener(CommandEvent.CONNECT, onConnect, 0)
	nc.AddEventListener(CommandEvent.CREATE_STREAM, onCreateStream, 0)
	nc.AddEventListener(CommandEvent.CLOSE, onDisconnect, 0)

	return nc, nil
}

func onConnect(e *CommandEvent.CommandEvent) {
	nc := e.Target.(*NetConnection.NetConnection)
	fmt.Printf("Application.onConnect: id=%s.\n", nc.FarID)

	var encoder AMF.Encoder
	var info *AMF.AMFObject
	if nc.ReadAccess == "/" || nc.ReadAccess == "/"+nc.AppName {
		Accept(nc)

		nc.SetAckWindowSize(2500000)
		nc.SetPeerBandwidth(2500000, LimitTypes.DYNAMIC)
		nc.SendUserControl(EventTypes.STREAM_BEGIN, 0, 0, 0)
		nc.SetChunkSize(4096)

		encoder.EncodeString(Commands.RESULT)
		info, _ = nc.GetInfoObject(Level.STATUS, Code.NETCONNECTION_CONNECT_SUCCESS, "connect success")
	} else {
		Reject(nc)

		encoder.EncodeString(Commands.ERROR)
		info, _ = nc.GetInfoObject(Level.ERROR, Code.NETCONNECTION_CONNECT_REJECTED, "connect reject")
	}

	encoder.EncodeNumber(1)
	encoder.EncodeObject(&rtmp.FMSProperties)

	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.DOUBLE,
		Key:  "objectEncoding",
		Data: float64(nc.ObjectEncoding),
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.ECMA_ARRAY,
		Key:  "data",
		Data: rtmp.FMSVersion,
	})
	encoder.EncodeObject(info)

	nc.SendEncodedBuffer(&encoder, e.Message.Header)
}

func onCreateStream(e *CommandEvent.CommandEvent) {
	nc := e.Target.(*NetConnection.NetConnection)
	ns, err := NetStream.New(nc)
	if err != nil {
		fmt.Printf("Failed to create NetStream: %v.\n", err)
		return
	}

	app, err := Get(nc.AppName)
	if err != nil {
		fmt.Printf("Failed to get application: %v.\n", err)
		return
	}

	ns.AddEventListener(CommandEvent.PUBLISH, app.onPublish, 0)
	ns.AddEventListener(CommandEvent.PLAY, app.onPlay, 0)
	ns.AddEventListener(CommandEvent.CLOSE, app.onUnpublish, 0)
}

func (this *Application) onPublish(e *CommandEvent.CommandEvent) {
	ns := e.Target.(*NetStream.NetStream)
	fmt.Printf("Application.onPublish: stream=%s.\n", ns.Stream.Name)

	if nc := ns.Nc; nc.WriteAccess == "/" || nc.WriteAccess == "/"+nc.AppName {
		inst, _ := this.GetInstance(nc.InstName)
		stream, _ := inst.GetStream(ns.Stream.Name, true)
		if stream == nil {
			info, _ := nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_FAILED, "Internal error")
			ns.SendStatus(e, info)
		} else if stream.Type == StreamTypes.PUBLISHING {
			info, _ := nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PUBLISH_BADNAME, "Publish bad name")
			ns.SendStatus(e, info)
		} else {
			ns.Stream.Type = StreamTypes.PUBLISHING
			ns.Stream.Sink(stream.Muxer)
			ns.Stream.Sink(stream.FMP4Muxer)

			info, _ := nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PUBLISH_START, "Publish start")
			ns.SendStatus(e, info)
		}
	} else {
		info, _ := nc.GetInfoObject(Level.ERROR, "No write access", "No write access")
		ns.SendStatus(e, info)
	}
}

func (this *Application) onPlay(e *CommandEvent.CommandEvent) {
	ns := e.Target.(*NetStream.NetStream)
	fmt.Printf("Application.onPlay: stream=%s.\n", ns.Stream.Name)

	if nc := ns.Nc; nc.ReadAccess == "/" || nc.ReadAccess == "/"+nc.AppName {
		inst, _ := this.GetInstance(nc.InstName)
		stream, _ := inst.GetStream(ns.Stream.Name, false)
		if stream != nil {
			nc.SendUserControl(EventTypes.STREAM_BEGIN, ns.Stream.ID, 0, 0)

			if e.Message.Reset {
				info, _ := nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_RESET, "Play reset")
				ns.SendStatus(e, info)
			}

			info, _ := nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_START, "Play start")
			ns.SendStatus(e, info)

			if stream.Type == StreamTypes.PLAYING_VOD {
				ns.Stream.Type = StreamTypes.PLAYING_VOD
			} else {
				ns.Stream.Type = StreamTypes.PLAYING_LIVE
			}

			switch nc.MuxerType {
			case MuxerTypes.RTMP:
				ns.Stream.Source(stream.Muxer)
			case MuxerTypes.FMP4:
				ns.Stream.Source(stream.FMP4Muxer)
			default:
			}
		} else {
			info, _ := nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PLAY_STREAMNOTFOUND, "Stream not found")
			ns.SendStatus(e, info)
		}
	} else {
		info, _ := nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PLAY_FAILED, "No read access")
		ns.SendStatus(e, info)
	}
}

func (this *Application) onUnpublish(e *CommandEvent.CommandEvent) {
	ns := e.Target.(*NetStream.NetStream)
	fmt.Printf("Application.onUnpublish: stream=%s.\n", ns.Stream.Name)

	inst, _ := this.GetInstance(ns.Nc.InstName)
	stream, _ := inst.GetStream(ns.Stream.Name, false)
	if stream != nil {
		stream.Muxer.EndOfStream("unpublish")
	}
}

func onDisconnect(e *CommandEvent.CommandEvent) {
	nc := e.Target.(*NetConnection.NetConnection)
	fmt.Printf("Application.onDisconnect: id=%s.\n", nc.FarID)

	Disconnect(nc)
}

func Accept(nc *NetConnection.NetConnection) (*Application, error) {
	app, err := Get(nc.AppName)
	if err != nil {
		return nil, err
	}

	inst, err := app.GetInstance(nc.InstName)
	inst.Add(nc)

	nc.Connected = true

	return app, nil
}

func Reject(nc *NetConnection.NetConnection) error {
	nc.Close()
	nc.Connected = false

	return nil
}

func Disconnect(nc *NetConnection.NetConnection) error {
	nc.Close()

	app, err := Get(nc.AppName)
	if err != nil {
		return err
	}

	inst, err := app.GetInstance(nc.InstName)
	inst.Remove(nc)

	return nil
}

func GC() {

}
