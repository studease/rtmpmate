package NetConnection

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"regexp"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/events/UserControlEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/AMF"
	AMFTypes "rtmpmate.com/net/rtmp/AMF/Types"
	"rtmpmate.com/net/rtmp/Chunk"
	"rtmpmate.com/net/rtmp/Chunk/CSIDs"
	"rtmpmate.com/net/rtmp/Chunk/States"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/AggregateMessage"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/BandwidthMessage"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage"
	EventTypes "rtmpmate.com/net/rtmp/Message/UserControlMessage/Event/Types"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/net/rtmp/ObjectEncoding"
	"rtmpmate.com/net/rtmp/Responder"
	"strconv"
	"syscall"
)

var (
	farID    int = 0
	urlRe, _     = regexp.Compile("^(rtmp[es]*)://([a-z0-9.-]+)(:([0-9]+))?/([a-z0-9.-_]+)(/([a-z0-9.-_]*))?$")
)

type NetConnection struct {
	conn              net.Conn
	wsConn            *websocket.Conn
	chunks            list.List
	farChunkSize      int
	nearChunkSize     int
	farAckWindowSize  uint32
	nearAckWindowSize uint32
	farBandwidth      uint32
	neerBandwidth     uint32
	farLimitType      byte
	neerLimitType     byte
	transactionID     int
	responders        map[int]*Responder.Responder

	Agent             string
	AppName           string
	AudioCodecs       uint64
	AudioSampleAccess string
	Connected         bool
	FarID             string
	InstName          string
	IP                string
	MuxerType         string
	NearID            string
	ObjectEncoding    byte
	PageURL           string
	Protocol          string
	ProtocolVersion   string
	ReadAccess        string
	Referrer          string
	Secure            bool
	URI               string
	VideoCodecs       uint64
	VideoSampleAccess string
	VirtualKey        string
	WriteAccess       string

	stats
	events.EventDispatcher
}

type stats struct {
	statsToAdmin

	pingRTT int

	audioQueueBytes int
	videoQueueBytes int
	soQueueBytes    int
	dataQueueBytes  int

	droppedAudioBytes int
	droppedVideoBytes int

	audioQueueMsgs int
	videoQueueMsgs int
	soQueueMsgs    int
	dataQueueMsgs  int

	droppedAudioMsgs int
	droppedVideoMsgs int
}

type statsToAdmin struct {
	connectTime float64

	bytesIn  uint32
	bytesOut uint32

	msgIn      int
	msgOut     int
	msgDropped int
}

func New(conn net.Conn) (*NetConnection, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	farID++

	var nc NetConnection
	nc.conn = conn
	nc.farChunkSize = 128
	nc.nearChunkSize = 128
	nc.transactionID = 0
	nc.responders = make(map[int]*Responder.Responder)

	nc.FarID = strconv.Itoa(farID)
	nc.InstName = "_definst_"
	nc.ObjectEncoding = ObjectEncoding.AMF0
	nc.ReadAccess = "/"
	nc.WriteAccess = "/"
	nc.AudioSampleAccess = "/"
	nc.VideoSampleAccess = "/"

	nc.AddEventListener(CommandEvent.CONNECT, nc.onConnect, 0)
	nc.AddEventListener(CommandEvent.CLOSE, nc.onClose, 0)
	nc.AddEventListener(CommandEvent.RESULT, nc.onResult, 0)
	nc.AddEventListener(CommandEvent.ERROR, nc.onError, 0)
	nc.AddEventListener(CommandEvent.CHECK_BANDWIDTH, nc.onCheckBandwidth, 0)
	nc.AddEventListener(CommandEvent.GET_STATS, nc.onGetStats, 0)

	return &nc, nil
}

func (this *NetConnection) Read(b []byte) (int, error) {
	return this.conn.Read(b)
}

func (this *NetConnection) Write(b []byte) (int, error) {
	switch this.Protocol {
	case "rtmp":
		return this.conn.Write(b)

	case "ws":
		err := this.wsConn.WriteMessage(websocket.BinaryMessage, b)
		if err != nil {
			return 0, err
		}

		return len(b), nil
	}

	return 0, fmt.Errorf("Unknown protocol: \"%s\".\n", this.Protocol)
}

func (this *NetConnection) WriteByChunk(b []byte, h *Message.Header) (int, error) {
	if h.Length < 2 {
		return 0, fmt.Errorf("chunk data (len=%d) not enough", h.Length)
	}

	var c Chunk.Chunk
	c.Fmt = h.Fmt

	for i := 0; i < h.Length; /* void */ {
		if h.CSID <= 63 {
			c.Data.WriteByte((c.Fmt << 6) | byte(h.CSID))
		} else if h.CSID <= 319 {
			c.Data.WriteByte((c.Fmt << 6) | 0x00)
			c.Data.WriteByte(byte(h.CSID - 64))
		} else if h.CSID <= 65599 {
			tmp := uint16(h.CSID)
			c.Data.WriteByte((c.Fmt << 6) | 0x01)
			err := binary.Write(&c.Data, binary.LittleEndian, &tmp)
			if err != nil {
				return i, err
			}
		} else {
			return i, fmt.Errorf("chunk size (%d) out of range", h.Length)
		}

		if c.Fmt <= 2 {
			if h.Timestamp >= 0xFFFFFF {
				c.Data.Write([]byte{0xFF, 0xFF, 0xFF})
			} else {
				c.Data.Write([]byte{
					byte(h.Timestamp>>16) & 0xFF,
					byte(h.Timestamp>>8) & 0xFF,
					byte(h.Timestamp>>0) & 0xFF,
				})
			}
		}
		if c.Fmt <= 1 {
			c.Data.Write([]byte{
				byte(h.Length>>16) & 0xFF,
				byte(h.Length>>8) & 0xFF,
				byte(h.Length>>0) & 0xFF,
			})
			c.Data.WriteByte(h.Type)
		}
		if c.Fmt == 0 {
			binary.Write(&c.Data, binary.LittleEndian, &h.StreamID)
		}

		// Extended Timestamp
		if h.Timestamp >= 0xFFFFFF {
			binary.Write(&c.Data, binary.BigEndian, &h.Timestamp)
		}

		// Chunk Data
		n := h.Length - i
		if n > this.nearChunkSize {
			n = this.nearChunkSize
		}

		_, err := c.Data.Write(b[i : i+n])
		if err != nil {
			return i, err
		}

		//fmt.Println(c.Data.Bytes())

		i += n

		if i < h.Length {
			switch h.Type {
			default:
				c.Fmt = 3
			}
		} else if i == h.Length {
			cs := c.Data.Bytes()
			_, err = this.Write(cs)
			if err != nil {
				return i, err
			}

			this.bytesOut += uint32(c.Data.Len())

			/*size := len(cs)
			for x := 0; x < size; x += 16 {
				fmt.Printf("\n")

				for y := 0; y < int(math.Min(float64(size-x), 16)); y++ {
					fmt.Printf("%02x ", cs[x+y])

					if y == 7 {
						fmt.Printf(" ")
					}
				}
			}*/
		} else {
			return i, fmt.Errorf("wrote too much")
		}
	}

	return h.Length, nil
}

func (this *NetConnection) Wait() error {
	var b = make([]byte, 14+4096)

	for {
		n, err := this.conn.Read(b)
		if err != nil {
			return err
		}

		this.bytesIn += uint32(n)

		err = this.parseChunk(b[:n], n)
		if err != nil {
			return err
		}
	}
}

func (this *NetConnection) WaitWebsocket(conn *websocket.Conn) error {
	this.wsConn = conn

	for {
		messageType, b, err := this.wsConn.ReadMessage()
		if err != nil {
			return err
		}

		n := len(b)
		this.bytesIn += uint32(n)

		switch messageType {
		case websocket.TextMessage:
			this.wsConn.WriteMessage(websocket.TextMessage, []byte("Not Acceptable"))

		case websocket.BinaryMessage:
			err = this.parseChunk(b, n)
			if err != nil {
				return err
			}

		case websocket.CloseMessage:
			this.Close()

		case websocket.PingMessage:
			// TODO:

		case websocket.PongMessage:
			// TODO:

		default:
			this.Close()
		}
	}
}

func (this *NetConnection) parseChunk(b []byte, size int) error {
	c := this.getUncompleteChunk()

	for i := 0; i < size; i++ {
		//tmp := uint32(b[i])
		//fmt.Printf("b[%d] = 0x%02x\n", i, tmp)

		switch c.State {
		case States.START:
			c.CurrentFmt = (b[i] >> 6) & 0xFF
			c.CSID = uint32(b[i]) & 0x3F

			if c.Polluted == false {
				c.Fmt = c.CurrentFmt
				c.Polluted = true
			}

			this.extendsFromPrecedingChunk(c)
			if c.CurrentFmt == 3 && c.Extended == false {
				c.State = States.DATA
			} else {
				c.State = States.FMT
			}

		case States.FMT:
			switch c.CSID {
			case 0:
				c.CSID = uint32(b[i]) + 64
				c.State = States.CSID_1
			case 1:
				c.CSID = uint32(b[i])
				c.State = States.CSID_0
			default:
				if c.CurrentFmt == 3 {
					if c.Extended {
						c.Timestamp = uint32(b[i]) << 24
						c.State = States.EXTENDED_TIMESTAMP_0
					} else {
						return fmt.Errorf("Failed to parse chunk: [1].")
					}
				} else {
					c.Timestamp = uint32(b[i]) << 16
					c.State = States.TIMESTAMP_0
				}
			}

		case States.CSID_0:
			c.CSID |= uint32(b[i]) << 8
			c.CSID += 64

			if c.CurrentFmt == 3 && c.Extended == false {
				c.State = States.DATA
			} else {
				c.State = States.CSID_1
			}

		case States.CSID_1:
			if c.CurrentFmt == 3 {
				if c.Extended {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [2].")
				}
			} else {
				c.Timestamp = uint32(b[i]) << 16
				c.State = States.TIMESTAMP_0
			}

		case States.TIMESTAMP_0:
			c.Timestamp |= uint32(b[i]) << 8
			c.State = States.TIMESTAMP_1

		case States.TIMESTAMP_1:
			c.Timestamp |= uint32(b[i])

			if c.CurrentFmt == 2 && c.Timestamp != 0xFFFFFF {
				c.State = States.DATA
			} else {
				c.State = States.TIMESTAMP_2
			}

		case States.TIMESTAMP_2:
			if c.CurrentFmt == 0 || c.CurrentFmt == 1 {
				c.MessageLength = int(b[i]) << 16
				c.State = States.MESSAGE_LENGTH_0
			} else if c.CurrentFmt == 2 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [3].")
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [4].")
			}

		case States.MESSAGE_LENGTH_0:
			c.MessageLength |= int(b[i]) << 8
			c.State = States.MESSAGE_LENGTH_1

		case States.MESSAGE_LENGTH_1:
			c.MessageLength |= int(b[i])
			c.State = States.MESSAGE_LENGTH_2

		case States.MESSAGE_LENGTH_2:
			c.MessageTypeID = b[i]

			if c.CurrentFmt == 1 && c.Timestamp != 0xFFFFFF {
				c.State = States.DATA
			} else {
				c.State = States.MESSAGE_TYPE_ID
			}

		case States.MESSAGE_TYPE_ID:
			if c.CurrentFmt == 0 {
				c.MessageStreamID = uint32(b[i])
				c.State = States.MESSAGE_STREAM_ID_0
			} else if c.CurrentFmt == 1 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [5].")
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [6].")
			}

		case States.MESSAGE_STREAM_ID_0:
			c.MessageStreamID |= uint32(b[i]) << 8
			c.State = States.MESSAGE_STREAM_ID_1

		case States.MESSAGE_STREAM_ID_1:
			c.MessageStreamID |= uint32(b[i]) << 16
			c.State = States.MESSAGE_STREAM_ID_2

		case States.MESSAGE_STREAM_ID_2:
			c.MessageStreamID |= uint32(b[i]) << 24
			if c.Timestamp == 0xFFFFFF {
				c.State = States.MESSAGE_STREAM_ID_3
			} else {
				c.State = States.DATA
			}

		case States.MESSAGE_STREAM_ID_3:
			if c.Timestamp == 0xFFFFFF {
				c.Timestamp = uint32(b[i]) << 24
				c.State = States.EXTENDED_TIMESTAMP_0
			} else {
				return fmt.Errorf("Failed to parse chunk: [7].")
			}

		case States.EXTENDED_TIMESTAMP_0:
			c.Timestamp |= uint32(b[i]) << 16
			c.State = States.EXTENDED_TIMESTAMP_1

		case States.EXTENDED_TIMESTAMP_1:
			c.Timestamp |= uint32(b[i]) << 8
			c.State = States.EXTENDED_TIMESTAMP_2

		case States.EXTENDED_TIMESTAMP_2:
			c.Timestamp |= uint32(b[i])
			c.State = States.EXTENDED_TIMESTAMP_3

		case States.EXTENDED_TIMESTAMP_3:
			c.State = States.DATA
			fallthrough
		case States.DATA:
			n := c.MessageLength - c.Data.Len()
			if n > size-i {
				n = size - i
			}
			if n > this.farChunkSize-c.Loaded {
				n = this.farChunkSize - c.Loaded
				c.Loaded = 0
				c.State = States.START
			} else {
				c.Loaded += n
			}

			_, err := c.Data.Write(b[i : i+n])
			if err != nil {
				return err
			}

			i += n - 1

			if c.Data.Len() < c.MessageLength {
				// c.State = States.DATA
			} else if c.Data.Len() == c.MessageLength {
				c.State = States.COMPLETE

				err := this.parseMessage(c)
				if err != nil {
					return err
				}

				if i < size-1 {
					c = this.getUncompleteChunk()
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [8].")
			}

		default:
			return fmt.Errorf("Failed to parse chunk: [9].")
		}
	}

	return nil
}

func (this *NetConnection) parseMessage(c *Chunk.Chunk) error {
	if c.MessageTypeID != 0x03 && c.MessageTypeID != 0x08 && c.MessageTypeID != 0x09 {
		fmt.Printf("\nonMessage: 0x%02x.\n", c.MessageTypeID)
	}

	b := c.Data.Bytes()
	size := c.Data.Len()

	switch c.MessageTypeID {
	case Types.SET_CHUNK_SIZE:
		this.farChunkSize = int(binary.BigEndian.Uint32(b) & 0x7FFFFFFF)
		fmt.Printf("Set farChunkSize: %d.\n", this.farChunkSize)

	case Types.ABORT:
		csid := binary.BigEndian.Uint32(b)
		fmt.Printf("Abort chunk stream: %d.\n", csid)

		element := this.chunks.Back()
		if element != nil {
			c := element.Value.(*Chunk.Chunk)
			if c.State != States.COMPLETE && c.CSID == csid {
				this.chunks.Remove(element)
				fmt.Printf("Removed uncomplete chunk %d.\n", csid)
			}
		}

	case Types.ACK:
		sequenceNumber := binary.BigEndian.Uint32(b)
		//fmt.Printf("Sequence Number: %d, Bytes out: %d.\n", sequenceNumber, this.bytesOut)

		if sequenceNumber != this.bytesOut {

		}

	case Types.USER_CONTROL:
		m, _ := UserControlMessage.New()
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onUserControl(m)
		if err != nil {
			return err
		}

	case Types.ACK_WINDOW_SIZE:
		this.farAckWindowSize = binary.BigEndian.Uint32(b)
		fmt.Printf("Set farAckWindowSize to %d.\n", this.farAckWindowSize)

	case Types.BANDWIDTH:
		m, _ := BandwidthMessage.New()
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onBandwidth(m)
		if err != nil {
			return err
		}

	case Types.EDGE:
		// TODO:

	case Types.AUDIO:
		m, _ := AudioMessage.New()
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, m))

	case Types.VIDEO:
		m, _ := VideoMessage.New()
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, m))

	case Types.AMF3_DATA:
		fallthrough
	case Types.DATA:
		m, _ := DataMessage.New(this.ObjectEncoding)
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(DataEvent.New(m.Handler, this, m))

	case Types.AMF3_SHARED_OBJECT:
		fallthrough
	case Types.SHARED_OBJECT:
		// TODO:

	case Types.AMF3_COMMAND:
		b = b[1:]
		fallthrough
	case Types.COMMAND:
		m, _ := CommandMessage.New(this.ObjectEncoding)
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		if m.CommandObject != nil {
			encoding, _ := m.CommandObject.Get("objectEncoding")
			if encoding != nil && encoding.Data.(float64) != 0 {
				this.ObjectEncoding = ObjectEncoding.AMF3
				m.Type = Types.AMF3_COMMAND
			}
		}

		err = this.onCommand(m)
		if err != nil {
			return err
		}

	case Types.AGGREGATE:
		m, _ := AggregateMessage.New()
		m.Fmt = c.Fmt
		m.CSID = c.CSID
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onAggregate(m)
		if err != nil {
			return err
		}

	default:
	}

	return nil
}

func (this *NetConnection) onUserControl(m *UserControlMessage.UserControlMessage) error {
	fmt.Printf("onUserControl: type=%d.\n", m.Event.Type)

	switch m.Event.Type {
	case EventTypes.STREAM_BEGIN:
		fmt.Printf("Stream Begin: id=%d.\n", m.Event.StreamID)

	case EventTypes.STREAM_EOF:
		fmt.Printf("Stream EOF: id=%d.\n", m.Event.StreamID)

	case EventTypes.STREAM_DRY:
		fmt.Printf("Stream Dry: id=%d.\n", m.Event.StreamID)

	case EventTypes.SET_BUFFER_LENGTH:
		fmt.Printf("Set BufferLength: id=%d, len=%dms.\n", m.Event.StreamID, m.Event.BufferLength)
		this.DispatchEvent(UserControlEvent.New(UserControlEvent.SET_BUFFER_LENGTH, this, m))

	case EventTypes.STREAM_IS_RECORDED:
		fmt.Printf("Stream is Recorded: id=%d.\n", m.Event.StreamID)

	case EventTypes.PING_REQUEST:
		fmt.Printf("Ping Request: timestamp=%d.\n", m.Event.Timestamp)

	case EventTypes.PING_RESPONSE:
		fmt.Printf("Ping Response: timestamp=%d.\n", m.Event.Timestamp)

	default:
	}

	return nil
}

func (this *NetConnection) onBandwidth(m *BandwidthMessage.BandwidthMessage) error {
	fmt.Printf("Set neerBandwidth: ack=%d, limit=%d.\n", m.AckWindowSize, m.LimitType)

	this.neerBandwidth = m.AckWindowSize
	this.neerLimitType = m.LimitType

	return nil
}

func (this *NetConnection) onCommand(m *CommandMessage.CommandMessage) error {
	fmt.Printf("onCommand: name=%s.\n", m.Name)

	if this.HasEventListener(m.Name) {
		this.DispatchEvent(CommandEvent.New(m.Name, this, m))
	} else {
		// Should not return error, this might be an user call
		fmt.Printf("No handler found for command \"%s\".\n", m.Name)
	}

	return nil
}

func (this *NetConnection) onConnect(e *CommandEvent.CommandEvent) {
	if this.Connected {
		fmt.Printf("Already connected.\n")
		return
	}

	// Init properties
	app, _ := e.Message.CommandObject.Get("app")
	if app != nil {
		this.AppName = app.Data.(string)
	}

	tcUrl, _ := e.Message.CommandObject.Get("tcUrl")
	if tcUrl != nil {
		arr := urlRe.FindStringSubmatch(tcUrl.Data.(string))
		if arr != nil {
			instName := arr[len(arr)-1]
			if instName != "" {
				this.InstName = instName
			}
		}
	}
}

func (this *NetConnection) onClose(e *CommandEvent.CommandEvent) {
	this.Connected = false
}

func (this *NetConnection) onResult(e *CommandEvent.CommandEvent) {

}

func (this *NetConnection) onError(e *CommandEvent.CommandEvent) {

}

func (this *NetConnection) onCheckBandwidth(e *Event.Event) {

}

func (this *NetConnection) onGetStats(e *Event.Event) {

}

func (this *NetConnection) onAggregate(m *AggregateMessage.AggregateMessage) error {
	fmt.Printf("onAggregate: id=%d, timstamp=%d, length=%d.\n", m.StreamID, m.Timestamp, m.Length)
	return nil
}

func (this *NetConnection) SendEncodedBuffer(encoder *AMF.Encoder, h Message.Header) error {
	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	h.Length = encoder.Len()
	this.WriteByChunk(b, &h)

	return nil
}

func (this *NetConnection) SetChunkSize(size int) error {
	var encoder AMF.Encoder
	encoder.AppendInt32(int32(size), false)

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	h.CSID = CSIDs.PROTOCOL_CONTROL
	h.Type = Types.SET_CHUNK_SIZE
	h.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &h)
	if err != nil {
		return err
	}

	this.nearChunkSize = size
	fmt.Printf("Set nearChunkSize: %d.\n", this.nearChunkSize)

	return nil
}

func (this *NetConnection) Abort() error {
	return nil
}

func (this *NetConnection) SendAckSequence() error {
	var encoder AMF.Encoder
	encoder.AppendInt32(int32(this.bytesIn), false)

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	h.CSID = CSIDs.PROTOCOL_CONTROL
	h.Type = Types.ACK
	h.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &h)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetConnection) SendUserControl(event uint16, streamID int, bufferLength int, timestamp int) error {
	var encoder AMF.Encoder
	encoder.AppendInt16(int16(event), false)
	if event <= EventTypes.STREAM_IS_RECORDED {
		encoder.AppendInt32(int32(streamID), false)
	}
	if event == EventTypes.SET_BUFFER_LENGTH {
		encoder.AppendInt32(int32(bufferLength), false)
	}
	if event == EventTypes.PING_REQUEST || event == EventTypes.PING_RESPONSE {
		encoder.AppendInt32(int32(timestamp), false)
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	m, _ := UserControlMessage.New()
	m.CSID = CSIDs.PROTOCOL_CONTROL
	m.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &m.Header)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetConnection) SetAckWindowSize(size int) error {
	var encoder AMF.Encoder
	encoder.AppendInt32(int32(size), false)

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	h.CSID = CSIDs.PROTOCOL_CONTROL
	h.Type = Types.ACK_WINDOW_SIZE
	h.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &h)
	if err != nil {
		return err
	}

	this.nearAckWindowSize = uint32(size)
	fmt.Printf("Set nearAckWindowSize: %d.\n", this.nearAckWindowSize)

	return nil
}

func (this *NetConnection) SetPeerBandwidth(size int, limitType byte) error {
	var encoder AMF.Encoder
	encoder.AppendInt32(int32(size), false)
	encoder.AppendInt8(int8(limitType))

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	m, _ := BandwidthMessage.New()
	m.CSID = CSIDs.PROTOCOL_CONTROL
	m.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &m.Header)
	if err != nil {
		return err
	}

	this.farBandwidth = uint32(size)
	this.farLimitType = limitType
	fmt.Printf("Set farBandwidth: ack=%d, limit=%d.\n", this.farBandwidth, this.farLimitType)

	return nil
}

func (this *NetConnection) Connect(uri string, args ...*AMF.AMFValue) error {
	return nil
}

func (this *NetConnection) CreateStream() error {
	return nil
}

func (this *NetConnection) Call(command string, responder *Responder.Responder, args ...*AMF.AMFValue) error {
	transactionID := 0
	if responder != nil {
		this.transactionID++
		transactionID = this.transactionID
	}

	var encoder AMF.Encoder
	encoder.EncodeString(command)
	encoder.EncodeNumber(float64(transactionID))
	for _, v := range args {
		encoder.EncodeValue(v)
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	h.CSID = CSIDs.COMMAND
	h.Type = Types.COMMAND
	h.Length = encoder.Len()

	_, err = this.WriteByChunk(b, &h)
	if err != nil {
		return err
	}

	if responder != nil {
		this.responders[transactionID] = responder
	}

	return nil
}

func (this *NetConnection) Close() error {
	err := this.conn.Close()

	if this.Connected {
		this.DispatchEvent(CommandEvent.New(CommandEvent.CLOSE, this, nil))
	}

	this.Connected = false

	return err
}

func (this *NetConnection) GetAppName() string {
	return this.AppName
}

func (this *NetConnection) GetInstName() string {
	return this.InstName
}

func (this *NetConnection) GetFarID() string {
	return this.FarID
}

func (this *NetConnection) GetInfoObject(level string, code string, description string) (*AMF.AMFObject, error) {
	var info AMF.AMFObject
	info.Init()

	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "level",
		Data: level,
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "code",
		Data: code,
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "description",
		Data: description,
	})

	return &info, nil
}

func (this *NetConnection) getUncompleteChunk() *Chunk.Chunk {
	for e := this.chunks.Back(); e != nil; e = e.Prev() {
		c := e.Value.(*Chunk.Chunk)
		if c.State != States.COMPLETE {
			return c
		}

		break
	}

	c, _ := Chunk.New()
	this.chunks.PushBack(c)

	return c
}

func (this *NetConnection) extendsFromPrecedingChunk(c *Chunk.Chunk) {
	if c.Fmt == 0 {
		return
	}

	for e, checking := this.chunks.Back(), false; e != nil; e = e.Prev() {
		b := e.Value.(*Chunk.Chunk)
		if b.CSID != c.CSID {
			continue
		}

		if checking == false {
			checking = true
			continue
		}

		if c.Fmt >= 1 && c.MessageStreamID == 0 {
			c.MessageStreamID = b.MessageStreamID
		}
		if c.Fmt >= 2 && c.MessageLength == 0 {
			c.MessageLength = b.MessageLength
			c.MessageTypeID = b.MessageTypeID
		}

		break
	}
}
