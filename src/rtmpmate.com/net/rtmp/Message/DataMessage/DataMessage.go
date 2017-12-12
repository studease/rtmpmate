package DataMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/AMF"
	AMFTypes "rtmpmate.com/net/rtmp/AMF/Types"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/ObjectEncoding"
)

type DataMessage struct {
	Message.Header
	Handler string
	Key     string
	Data    *AMF.AMFObject
	Payload []byte
}

func New(encoding byte) (*DataMessage, error) {
	var m DataMessage

	if encoding == ObjectEncoding.AMF0 {
		m.Type = Types.DATA
	} else {
		m.Type = Types.AMF3_DATA
	}

	return &m, nil
}

func (this *DataMessage) Parse(b []byte, offset int, size int) error {
	cost := 0

	v, err := AMF.DecodeValue(b, offset+cost, size-cost)
	if err != nil {
		return err
	}

	cost += v.Cost
	this.Handler = v.Data.(string)

	v, err = AMF.DecodeValue(b, offset+cost, size-cost)
	if err != nil {
		return err
	}

	cost += v.Cost
	this.Key = v.Data.(string)

	v, _ = AMF.DecodeValue(b, offset+cost, size-cost)
	if v != nil {
		cost += v.Cost
		this.Data = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

		if v.Type == AMFTypes.OBJECT || v.Type == AMFTypes.ECMA_ARRAY || v.Type == AMFTypes.STRICT_ARRAY {
			this.Data.Data = v.Data.(list.List)
		}
	}

	this.Payload = b[offset : offset+size]

	return nil
}
