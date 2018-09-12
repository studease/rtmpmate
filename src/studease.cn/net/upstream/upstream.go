package upstream

import (
	"errors"
	"fmt"
	"strings"
	"studease.cn/utils"
	"studease.cn/utils/key"
	"sync"
	"sync/atomic"
)

const (
	FLAG_NORMAL = 0x00
	FLAG_BACKUP = 0x01
	FLAG_DOWN   = 0x02
)

type Server struct {
	Name     string
	Address  string
	Port     uint16
	Flag     int32
	Weight   int32
	Timeout  int32
	MaxFails int32

	sync.RWMutex
}

type Upstream struct {
	server []*Server
	index  int
	alive  int
	backup int

	sync.RWMutex
}

var (
	mtx       sync.RWMutex
	upstreams map[string]*Upstream

	flags = map[string]int32{
		"normal": FLAG_NORMAL,
		"backup": FLAG_BACKUP,
		"down":   FLAG_DOWN,
	}
)

func init() {
	upstreams = make(map[string]*Upstream)
}

func (this *Server) Init(cnf utils.Conf) *Server {
	var (
		v  interface{}
		ok bool
	)

	*this = Server{Name: "origin", Address: "127.0.0.1", Port: 80, Weight: 5, Timeout: 5}

	if v, ok = cnf[key.NAME]; ok {
		this.Name = v.(string)
	}

	if v, ok = cnf[key.ADDRESS]; ok {
		this.Address = v.(string)
	}

	if v, ok = cnf[key.PORT]; ok {
		this.Port = uint16(v.(float64))
	}

	if v, ok = cnf[key.FLAG]; ok {
		arr := strings.Split(v.(string), "|")
		for _, name := range arr {
			this.Flag |= flags[name]
		}
	}

	if v, ok = cnf[key.WEIGHT]; ok {
		this.Weight = int32(v.(float64))
	}

	if v, ok = cnf[key.TIMEOUT]; ok {
		this.Timeout = int32(v.(float64))
	}

	if v, ok = cnf[key.MAX_FAILS]; ok {
		this.MaxFails = int32(v.(float64))
	}

	return this
}

func (this *Server) URL(cnf utils.Conf, port uint16) string {
	this.RLock()

	url := cnf[key.PROTOCOL].(string) + this.Address
	if this.Port != port {
		url += fmt.Sprintf(":%d", this.Port)
	}

	url += cnf[key.URL].(string)

	this.RUnlock()

	return url
}

func (this *Upstream) Init() *Upstream {
	return this
}

func (this *Upstream) Add(s *Server) {
	this.Lock()

	this.server = append(this.server, s)

	if s.Flag == FLAG_BACKUP {
		this.backup++
	} else if s.Flag == FLAG_NORMAL {
		this.alive++
	}

	this.Unlock()
}

func (this *Upstream) Get() (*Server, error) {
	var (
		s   *Server = nil
		err error
	)

	this.RLock()

	if this.alive > 0 || this.backup > 0 {
		for {
			s = this.server[this.index]

			if this.index++; this.index >= len(this.server) {
				this.index = 0
			}

			if this.alive > 0 {
				if atomic.LoadInt32(&s.Flag) == FLAG_NORMAL {
					break
				}
			} else if this.backup > 0 {
				if atomic.LoadInt32(&s.Flag) == FLAG_BACKUP {
					break
				}
			}
		}
	} else {
		err = errors.New("no server usable")
	}

	this.RUnlock()

	return s, err
}

func Add(arr []interface{}) {
	mtx.Lock()

	for _, v := range arr {
		s := new(Server).Init(utils.Conf(v.(map[string]interface{})))

		u, ok := upstreams[s.Name]
		if ok == false {
			u = new(Upstream)
			upstreams[s.Name] = u
		}

		u.Add(s)
	}

	mtx.Unlock()
}

func Get(name string) (*Server, error) {
	mtx.RLock()

	u, ok := upstreams[name]
	if ok == false {
		return nil, errors.New("upstream not found")
	}

	mtx.RUnlock()

	return u.Get()
}
