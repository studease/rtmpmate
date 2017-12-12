package Body

import (
	"rtmpmate.com/net/rtmp/Message"
)

type Body struct {
	Message.Message
	Size int
}

func New() (*Body, error) {
	var b Body
	return &b, nil
}
