package channel

import (
	"encoding/json"
	"fmt"
	"sort"
	"studease.cn/events/TimerEvent"
	"studease.cn/net/chat/channel/group"
	"studease.cn/net/chat/channel/sanction"
	"studease.cn/net/chat/message"
	"studease.cn/net/chat/user"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"studease.cn/utils/timer"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_GROUPS = 2
)

const (
	FLAG_RECORD = 0x01
)

var (
	mtx        sync.RWMutex
	channels   map[string]*Channel
	PushPeriod = time.Duration(30) * time.Second
	PushEnable = true
)

func init() {
	channels = make(map[string]*Channel)
}

type Info struct {
	ID    string `json:"id"`
	Group int32  `json:"group"` // used on cmd=info
	Total int32  `json:"total"` // used on cmd=info
	Stat  int32  `json:"stat"`
}

type Channel struct {
	timer     *timer.Timer
	groups    int32
	capacity  int32
	table     map[int32]*group.Group
	array     []*group.Group
	length    int32
	sanctions map[string]*sanction.Sanction
	Flag      int32

	Info
	sync.RWMutex
}

func New(id string, groups int32, capacity int32) *Channel {
	return new(Channel).Init(id, groups, capacity)
}

func (this *Channel) Init(id string, groups int32, capacity int32) *Channel {
	this.ID = id
	this.groups = groups
	this.capacity = capacity

	if groups == 0 { // Auto grow
		groups = DEFAULT_GROUPS
	}

	this.table = make(map[int32]*group.Group, groups)
	this.array = make([]*group.Group, 0, groups)
	this.sanctions = make(map[string]*sanction.Sanction)

	log.Debug(level.CHAT, "init Channel(%s, %d, %d)", this.ID, this.groups, this.capacity)

	if PushEnable {
		this.timer = timer.New(PushPeriod, 0)
		this.timer.AddEventListener(TimerEvent.TIMER, this.onTimer, 0)
		this.timer.Start()
	}

	return this
}

func (this *Channel) Get(id int32) (*group.Group, bool) {
	this.RLock()
	g, ok := this.table[id]
	this.RUnlock()

	return g, ok
}

func (this *Channel) Add(usr *user.User) (*group.Group, error) {
	var g *group.Group

	this.Lock()
	this.sort()

	n := int32(len(this.array))
	if n > 0 {
		g = this.array[0]
	}

	if g == nil || g.IsFull() {
		if this.groups > 0 && this.groups == n {
			err := fmt.Errorf("max user amount reached: cid=%s", this.ID)
			log.Warn("%v", err)
			this.Unlock()
			return nil, err
		}

		g = group.New(n, this.capacity)

		this.table[g.ID] = g
		this.array = append(this.array, g)

		log.Debug(level.CHAT, "new group: cid=%s, gid=%d", this.ID, g.ID)
	}

	usr.Channel = this.ID

	this.Unlock()

	err := g.Add(usr)

	if err == nil {
		atomic.AddInt32(&this.length, 1)
		log.Debug(level.CHAT, "add user: cid=%s, uid=%s", this.ID, usr.ID)
	} else {
		log.Debug(level.CHAT, "failed to add user: cid=%s, uid=%s, err=%v", this.ID, usr.ID, err)
	}

	return g, err
}

func (this *Channel) Remove(usr *user.User) bool {
	this.RLock()
	g, ok := this.table[usr.Group]
	this.RUnlock()

	if !ok {
		return false
	}

	ok = g.Remove(usr)
	if ok {
		atomic.AddInt32(&this.length, -1)
		log.Debug(level.CHAT, "remove user: cid=%s, uid=%s", this.ID, usr.ID)
	}

	return ok
}

func (this *Channel) Find(usr string, group int32) (*user.User, bool) {
	this.RLock()
	g, ok := this.table[group]
	this.RUnlock()

	if !ok {
		return nil, false
	}

	return g.Find(usr)
}

func (this *Channel) Freeze(usr *user.User, sub string, opt string, d time.Duration) *sanction.Element {
	this.Lock()

	s, ok := this.sanctions[sub]
	if !ok {
		s = sanction.New()
		this.sanctions[sub] = s
	}

	e := s.Add(usr, opt, d)

	this.Unlock()

	return e
}

func (this *Channel) Unfreeze(sub string, opt string) {
	this.Lock()

	s, ok := this.sanctions[sub]
	if ok {
		s.Remove(opt)
		if s.Len() == 0 {
			delete(this.sanctions, sub)
		}
	}

	this.Unlock()
}

func (this *Channel) Limited(sub string, opt string) bool {
	this.Lock()

	s, ok := this.sanctions[sub]
	if ok {
		ok = s.Limited(opt)
		if !ok && s.Len() == 0 {
			delete(this.sanctions, sub)
		}
	}

	this.Unlock()

	return ok
}

func (this *Channel) Foreach(cb func(id string, usr *user.User)) {
	this.RLock()

	for _, g := range this.array {
		g.Foreach(cb)
	}

	this.RUnlock()
}

func (this *Channel) Length() int32 {
	return atomic.LoadInt32(&this.length)
}

func (this *Channel) onTimer(e *TimerEvent.TimerEvent) {
	this.RLock()

	raw := map[string]interface{}{
		message.KEY_CMD:  message.CMD_USER,
		message.KEY_DATA: map[string]interface{}{},
		message.KEY_MODE: message.MODE_BROADCAST,
		message.KEY_CHANNEL: map[string]interface{}{
			message.KEY_ID:    this.ID,
			message.KEY_TOTAL: this.length,
			message.KEY_STAT:  this.Stat,
		},
	}

	this.RUnlock()

	data, err := json.Marshal(raw)
	if err != nil {
		log.Warn("failed to marshal message: cmd=%s", message.CMD_USER)
		return
	}

	this.Foreach(func(id string, usr *user.User) {
		usr.Send(data)
	})
}

func (this *Channel) sort() {
	sort.Sort(this)
}

func (this *Channel) Len() int {
	return len(this.array)
}

func (this *Channel) Less(i, j int) bool {
	return this.array[j].Len() < this.array[i].Len()
}

func (this *Channel) Swap(i, j int) {
	this.array[i], this.array[j] = this.array[j], this.array[i]
}

func Get(id string) (*Channel, bool) {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()
	return ch, ok
}

func Add(id string, usr *user.User, groups int32, capacity int32) (*Channel, *group.Group, error) {
	mtx.Lock()

	ch, ok := channels[id]
	if !ok {
		ch = New(id, groups, capacity)
		channels[id] = ch
	}

	mtx.Unlock()

	g, err := ch.Add(usr)

	return ch, g, err
}

func Remove(id string, usr *user.User) bool {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()

	if !ok {
		return false
	}

	return ch.Remove(usr)
}

func Freeze(id string, usr *user.User, sub string, opt string, d time.Duration) *sanction.Element {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()

	if !ok {
		return nil
	}

	return ch.Freeze(usr, sub, opt, d)
}

func Unfreeze(id string, sub string, opt string) {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()

	if ok {
		ch.Unfreeze(sub, opt)
	}
}

func Limited(id string, usr string, opt string) bool {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()

	if ok {
		ok = ch.Limited(usr, opt)
	}

	return ok
}

func Foreach(id string, cb func(id string, usr *user.User)) {
	mtx.RLock()
	ch, ok := channels[id]
	mtx.RUnlock()

	if ok {
		ch.Foreach(cb)
	}
}
