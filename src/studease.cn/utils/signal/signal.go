package signal

import (
	"studease.cn/events"
	"studease.cn/events/SignalEvent"
	"sync"
)

const (
	EXIT = 0
)

var (
	fd = New()
)

type Signal struct {
	C chan int

	sync.RWMutex
	events.EventDispatcher
}

func New() *Signal {
	return new(Signal).Init()
}

func (this *Signal) Init() *Signal {
	this.C = make(chan int)
	return this
}

func (this *Signal) Send(n int) {
	this.Lock()
	this.C <- n
	this.Unlock()
}

func (this *Signal) Wait() {
	for {
		n := <-this.C

		this.RLock()
		this.DispatchEvent(SignalEvent.New(SignalEvent.SIGNAL, this, n))
		this.RUnlock()
	}
}

func Default() *Signal {
	return fd
}
