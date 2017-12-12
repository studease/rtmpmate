package CommandMessage

import (
	"container/list"
	"fmt"
	"math"
	"rtmpmate.com/net/rtmp/AMF"
	AMFTypes "rtmpmate.com/net/rtmp/AMF/Types"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/ObjectEncoding"
)

type CommandMessage struct {
	Message.Header
	Name          string
	TransactionID uint64

	CommandObject  *AMF.AMFObject
	Arguments      *AMF.AMFObject
	Response       *AMF.AMFValue
	StreamID       uint64
	StreamName     string
	Start          float64
	Duration       float64
	Reset          bool
	Flag           bool
	PublishingName string
	PublishingType string
	MilliSeconds   float64
	Pause          bool
}

func New(encoding byte) (*CommandMessage, error) {
	var m CommandMessage

	if encoding == ObjectEncoding.AMF0 {
		m.Type = Types.COMMAND
	} else {
		m.Type = Types.AMF3_COMMAND
	}

	return &m, nil
}

func (this *CommandMessage) Parse(b []byte, offset int, size int) error {
	cost := 0

	v, err := AMF.DecodeValue(b, offset+cost, size-cost)
	if err != nil {
		return err
	}

	cost += v.Cost
	this.Name = v.Data.(string)

	v, err = AMF.DecodeValue(b, offset+cost, size-cost)
	if err != nil {
		return err
	}

	cost += v.Cost
	this.TransactionID = math.Float64bits(v.Data.(float64))

	switch this.Name {
	// NetConnection Commands
	case Commands.CONNECT:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Data:    v.Data.(list.List),
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

		v, _ = AMF.DecodeValue(b, offset+cost, size-cost)
		if v != nil {
			cost += v.Cost
			this.Arguments = &AMF.AMFObject{
				AMFHash: AMF.AMFHash{v.Hash},
				Data:    v.Data.(list.List),
				Cost:    v.Cost,
				Ended:   v.Ended,
			}
		}

	case Commands.CLOSE:
		fmt.Printf("Parsing command %s.\n", this.Name)

	case Commands.CREATE_STREAM:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Cost:    v.Cost,
			Ended:   v.Ended,
		}
		if v.Type == AMFTypes.OBJECT || v.Type == AMFTypes.ECMA_ARRAY || v.Type == AMFTypes.STRICT_ARRAY {
			this.CommandObject.Data = v.Data.(list.List)
		}

	case Commands.RESULT:
		fallthrough
	case Commands.ERROR:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Cost:    v.Cost,
			Ended:   v.Ended,
		}
		if v.Type == AMFTypes.OBJECT || v.Type == AMFTypes.ECMA_ARRAY || v.Type == AMFTypes.STRICT_ARRAY {
			this.CommandObject.Data = v.Data.(list.List)
		}

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.Response = v

	// NetStream Commands
	case Commands.PLAY:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.StreamName = v.Data.(string)

		v, _ = AMF.DecodeValue(b, offset+cost, size-cost)
		if v == nil {
			this.Start = -2
		} else {
			cost += v.Cost
			this.Start = v.Data.(float64)
		}

		v, _ = AMF.DecodeValue(b, offset+cost, size-cost)
		if v == nil {
			this.Duration = -1
		} else {
			cost += v.Cost
			this.Duration = v.Data.(float64)
		}

		v, _ = AMF.DecodeValue(b, offset+cost, size-cost)
		if v == nil {
			this.Reset = true
		} else {
			cost += v.Cost
			this.Reset = v.Data.(bool)
		}

	case Commands.PLAY2:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.Arguments = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Data:    v.Data.(list.List),
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

	case Commands.DELETE_STREAM:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.StreamID = math.Float64bits(v.Data.(float64))

	case Commands.CLOSE_STREAM:
		fmt.Printf("Parsing command %s.\n", this.Name)

	case Commands.RECEIVE_AUDIO:
		fallthrough
	case Commands.RECEIVE_VIDEO:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.Flag = v.Data.(bool)

	case Commands.PUBLISH:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.PublishingName = v.Data.(string)

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.PublishingType = v.Data.(string)

	case Commands.SEEK:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.MilliSeconds = v.Data.(float64)

	case Commands.PAUSE:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.Pause = v.Data.(bool)

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.MilliSeconds = v.Data.(float64)

	case Commands.ON_STATUS:
		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset+cost, size-cost)
		if err != nil {
			return err
		}

		cost += v.Cost
		this.Response = v

	default:
	}

	return nil
}
