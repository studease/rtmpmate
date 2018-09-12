package timer

import (
	"studease.cn/events"
	"studease.cn/events/TimerEvent"
	"sync"
	"time"
)

type Timer struct {
	ticker       *time.Ticker
	running      bool
	Delay        time.Duration
	RepeatCount  int32
	CurrentCount int32

	sync.RWMutex
	events.EventDispatcher
}

func New(delay time.Duration, repeatCount int32) *Timer {
	return new(Timer).Init(delay, repeatCount)
}

func (this *Timer) Init(delay time.Duration, repeatCount int32) *Timer {
	this.Delay = delay
	this.RepeatCount = repeatCount
	return this
}

func (this *Timer) Start() {
	this.Lock()

	if !this.running {
		this.ticker = time.NewTicker(this.Delay)
		this.running = true

		go this.wait()
	}

	this.Unlock()
}

func (this *Timer) wait() {
	for {
		select {
		case <-this.ticker.C:
			this.Lock()

			this.CurrentCount++
			this.DispatchEvent(TimerEvent.New(TimerEvent.TIMER, this))

			if this.RepeatCount > 0 && this.CurrentCount == this.RepeatCount {
				this.stop()
				this.DispatchEvent(TimerEvent.New(TimerEvent.TIMER_COMPLETE, this))
			}

			this.Unlock()
		}
	}
}

func (this *Timer) stop() {
	if this.ticker != nil {
		this.ticker.Stop()
	}

	this.running = false
}

func (this *Timer) Stop() {
	this.Lock()
	this.stop()
	this.Unlock()
}

func (this *Timer) Reset() {
	this.Lock()
	this.stop()
	this.CurrentCount = 0
	this.Unlock()
}

func (this *Timer) Running() bool {
	this.RLock()
	b := this.running
	this.RUnlock()
	return b
}
