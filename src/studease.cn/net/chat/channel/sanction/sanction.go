package sanction

import (
	"studease.cn/net/chat/user"
	"sync"
	"time"
)

type Element struct {
	Manager user.Info
	Until   time.Time
}

func (this *Element) Update(mgr *user.Info, d time.Duration) *Element {
	this.Manager = *mgr
	this.Until = time.Now().Add(d)
	return this
}

type Sanction struct {
	table map[string]*Element
	sync.RWMutex
}

func New() *Sanction {
	return new(Sanction).Init()
}

func (this *Sanction) Init() *Sanction {
	this.table = make(map[string]*Element)
	return this
}

func (this *Sanction) Add(usr *user.User, opt string, d time.Duration) *Element {
	this.Lock()

	e, ok := this.table[opt]
	if ok == false {
		e = new(Element)
		this.table[opt] = e
	}

	e.Update(&usr.Info, d)

	this.Unlock()

	return e
}

func (this *Sanction) Remove(opt string) {
	this.Lock()
	delete(this.table, opt)
	this.Unlock()
}

func (this *Sanction) Limited(opt string) bool {
	this.Lock()

	e, ok := this.table[opt]
	if ok {
		ok = time.Now().Before(e.Until)
		if !ok {
			delete(this.table, opt)
		}
	}

	this.Unlock()

	return ok
}

func (this *Sanction) Len() int {
	this.RLock()
	n := len(this.table)
	this.RUnlock()
	return n
}
