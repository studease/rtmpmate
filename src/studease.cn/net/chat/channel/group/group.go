package group

import (
	"container/list"
	"studease.cn/net/chat/user"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"sync"
	"sync/atomic"
)

const (
	DEFAULT_CAPACITY = 1000
)

type Group struct {
	ID       int32
	capacity int32
	table    map[string]*list.List
	length   int32

	sync.RWMutex
}

func New(id int32, capacity int32) *Group {
	return new(Group).Init(id, capacity)
}

func (this *Group) Init(id int32, capacity int32) *Group {
	this.ID = id
	this.capacity = capacity

	if capacity == 0 { // Auto grow
		capacity = DEFAULT_CAPACITY
	}

	this.table = make(map[string]*list.List, capacity)

	log.Debug(level.CHAT, "init Group(%d, %d)", this.ID, this.capacity)

	return this
}

func (this *Group) Add(usr *user.User) error {
	id := usr.ID

	this.Lock()

	l, ok := this.table[id]
	if ok == false {
		l = list.New()
		this.table[id] = l
	}

	l.PushBack(usr)
	atomic.AddInt32(&this.length, 1)

	usr.Group = this.ID

	this.Unlock()

	return nil
}

func (this *Group) Remove(usr *user.User) bool {
	id := usr.ID

	this.Lock()

	l, ok := this.table[id]
	if !ok {
		this.Unlock()
		return false
	}

	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == usr {
			l.Remove(e)
			atomic.AddInt32(&this.length, -1)
			this.Unlock()
			return true
		}
	}

	this.Unlock()

	return false
}

func (this *Group) Find(id string) (*user.User, bool) {
	var usr *user.User

	this.RLock()

	l, ok := this.table[id]
	if ok {
		if l.Len() > 0 {
			usr = l.Front().Value.(*user.User)
		} else {
			ok = false
		}
	}

	this.RUnlock()

	return usr, ok
}

func (this *Group) Foreach(cb func(id string, usr *user.User)) {
	this.RLock()

	for k, l := range this.table {
		for e := l.Front(); e != nil; e = e.Next() {
			cb(k, e.Value.(*user.User))
		}
	}

	this.RUnlock()
}

func (this *Group) Len() int {
	return int(atomic.LoadInt32(&this.length))
}

func (this *Group) IsFull() bool {
	capacity := atomic.LoadInt32(&this.capacity)
	length := atomic.LoadInt32(&this.length)
	return capacity > 0 && capacity <= length
}
