package license

import (
	"studease.cn/events"
	"studease.cn/events/Event"
)

type License struct {
	events.EventDispatcher
}

func New() (*License, error) {
	var lic License
	return &lic, nil
}

func (this *License) Check(s string) {
	this.DispatchEvent(Event.New(Event.COMPLETE, this))
}
