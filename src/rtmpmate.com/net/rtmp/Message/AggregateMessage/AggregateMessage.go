package AggregateMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/AggregateMessage/Body"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AggregateMessage struct {
	Message.Header
	Body list.List
}

func New() (*AggregateMessage, error) {
	var m AggregateMessage
	m.Type = Types.AGGREGATE

	return &m, nil
}

func (this *AggregateMessage) Parse(b []byte, offset int, size int) error {
	m, _ := Message.New()

	err := m.Parse(b, offset, size)
	if err != nil {
		return err
	}

	this.Length = m.Length
	this.Timestamp = m.Timestamp
	this.StreamID = m.StreamID

	var body *Body.Body
	for i := 0; i < m.Length; i += body.Size {
		body, err := Body.New()
		if err != nil {
			return err
		}

		err = body.Parse(m.Payload, i, m.Length)
		if err != nil {
			return err
		}

		this.Body.PushBack(body)
	}

	return nil
}
