package chat

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"studease.cn/events/ChatEvent"
	"studease.cn/events/Event"
	"studease.cn/net/chat/channel"
	"studease.cn/net/chat/channel/group"
	"studease.cn/net/chat/message"
	"studease.cn/net/chat/user"
	"studease.cn/net/chat/user/role"
	httpx "studease.cn/net/http"
	"studease.cn/net/upstream"
	"studease.cn/utils"
	"studease.cn/utils/key"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"studease.cn/utils/register"
	"sync/atomic"
	"time"
)

const (
	CHAT_HANDLER = "chat"
)

const (
	_CAPACITY   = "capacity"
	_GROUP      = "group"
	_ICON       = "icon"
	_PERIOD     = "period"
	_MAXIMUM    = "maximum"
	_ON_CONTROL = "on_control"
	_ROLE       = "role"
	_SEED       = "seed"
	_USER       = "user"
	_VISITOR    = "visitor"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	defaults = utils.Conf{
		"protocol": "",
		"on_connect": utils.Conf{
			"protocol": "http",
			"method":   "GET",
			"upstream": "web",
			"path":     "",
			"args":     "token",
			"enable":   false,
		},
		"on_control": utils.Conf{
			"protocol": "http://",
			"method":   "GET",
			"upstream": "web",
			"url":      "/data/userctrl.json",
			"args":     "",
			"enable":   false,
		},
		"visitor": utils.Conf{
			"seed":   999,
			"name":   "游客%03d",
			"icon":   "",
			"role":   0,
			"enable": true,
		},
		"group": utils.Conf{
			"length":   0,
			"capacity": 1000,
		},
		"users": utils.Conf{
			"interval": 30,
			"enable":   true,
		},
	}
)

func init() {
	register.Add(CHAT_HANDLER, reflect.ValueOf(ChatHandler{}).Type())
	rand.Seed(time.Now().UnixNano())
}

type IdentResponse struct {
	User    user.Info    `json:"user"`
	Channel channel.Info `json:"channel"`
}

type CtrlResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}

type ChatHandler struct {
	server *httpx.Server
	conf   utils.Conf
}

func (this *ChatHandler) Init(srv *httpx.Server, cnf utils.Conf) {
	this.server = srv
	this.conf = utils.Extends(make(utils.Conf), defaults, cnf)

	upgrader.CheckOrigin = this.checkOrigin

	tmp := this.conf[_USER].(utils.Conf)
	channel.PushPeriod = time.Duration(tmp[_PERIOD].(float64)) * time.Second
	channel.PushEnable = tmp[key.ENABLE].(bool)
}

func (this *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		info *IdentResponse
		conn *websocket.Conn
		ch   *channel.Channel
		g    *group.Group
		err  error
	)

	w.Header().Set("Server", httpx.SERVER_NAME)

	if !websocket.IsWebSocketUpgrade(r) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	info, err = this.ident(w, r)
	if err != nil {
		log.Debug(level.CHAT, "%s", err)
		return
	}

	header := make(http.Header)
	header.Set("Server", httpx.SERVER_NAME)
	header.Set("Access-Control-Allow-Origin", this.server.Cors)
	header.Set("Sec-Websocket-Protocol", this.conf[key.PROTOCOL].(string))

	conn, err = upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Debug(level.CHAT, "%s", err)
		return
	}

	usr := user.New(conn, &info.User)
	usr.AddEventListener(ChatEvent.MESSAGE, this.onMessage, 0)
	usr.AddEventListener(Event.CLOSE, this.onClose, 0)

	cnf := this.conf[_GROUP].(utils.Conf)
	groups := int32(cnf[_MAXIMUM].(float64))
	capacity := int32(cnf[_CAPACITY].(float64))

	ch, g, err = channel.Add(info.Channel.ID, usr, groups, capacity)
	if err != nil {
		log.Debug(level.CHAT, "%s", err)
		usr.Close()
		return
	}

	info.Channel.Group = atomic.LoadInt32(&g.ID)
	info.Channel.Total = ch.Length()

	raw := map[string]interface{}{
		message.KEY_CMD:     message.CMD_INFO,
		message.KEY_USER:    info.User,
		message.KEY_CHANNEL: info.Channel,
	}

	err = usr.SendRaw(raw)
	if err != nil {
		log.Debug(level.CHAT, "%s", err)
		usr.Close()
		return
	}

	err = usr.Read()
	if err != nil {
		log.Debug(level.CHAT, "%s", err)
	}

	usr.Close()
}

func (this *ChatHandler) ident(w http.ResponseWriter, r *http.Request) (*IdentResponse, error) {
	var (
		data IdentResponse
		req  *http.Request
		res  *http.Response
		body []byte
	)

	cid := this.getChannelId(r.URL.Path)
	cnf := this.conf[key.ON_CONNECT].(utils.Conf)

	args := cnf[key.ARGS].(string)
	keys := strings.Split(args, "|")

	src := r.URL.Query()

	if token := src.Get(keys[0]); token != "" && cnf[key.ENABLE].(bool) {
		s, err := upstream.Get(cnf[key.UPSTREAM].(string))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		req, err = http.NewRequest(cnf[key.METHOD].(string), s.URL(cnf, 80), nil)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		dst := req.URL.Query()
		dst.Add("channel", cid)

		if args != "" {
			for _, key := range keys {
				dst.Add(key, src.Get(key))
			}
		}

		req.URL.RawQuery = dst.Encode()

		client := new(http.Client)

		res, err = client.Do(req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return nil, err
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if err != nil {
			log.Debug(level.CHAT, "%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Debug(level.CHAT, "failed to unmarshal ident response: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return nil, err
		}
	} else {
		cnf, _ = this.conf[_VISITOR].(utils.Conf)
		if cnf[key.ENABLE].(bool) == false {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, fmt.Errorf("handler \"%s\" not allowed", CHAT_HANDLER)
		}

		seed, _ := cnf[_SEED]
		id := rand.Intn(int(seed.(float64)))

		data.User.ID = fmt.Sprintf("%d", id)
		data.User.Name = fmt.Sprintf(cnf[key.NAME].(string), id)
		data.User.Icon = cnf[_ICON].(string)
		data.User.Role = int32(cnf[_ROLE].(float64))

		data.Channel.ID = cid
		data.Channel.Stat = 0
	}

	return &data, nil
}

func (this *ChatHandler) onMessage(e *ChatEvent.ChatEvent) {
	usr := e.Target.(*user.User)

	ch, ok := channel.Get(usr.Channel)
	if !ok {
		usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	raw := map[string]interface{}{
		message.KEY_CMD:  e.Cmd,
		message.KEY_DATA: e.Data,
		message.KEY_MODE: e.Mode,
		message.KEY_USER: usr.Info,
	}

	if e.Seq != 0 {
		raw[message.KEY_SEQ] = e.Seq
	}

	switch e.Cmd {
	case message.CMD_TEXT:
		if usr.Role < atomic.LoadInt32(&ch.Stat) || ch.Limited(usr.ID, message.OPT_MUTE) {
			usr.Error(http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

	case message.CMD_CTRL:
		opt, d := message.ParseControl(e.Data)

		_, err := this.control(usr, usr.Channel, e.Sub, opt, d)
		if err != nil {
			usr.Error(http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		if d == time.Duration(0) {
			ch.Unfreeze(e.Sub, opt)
		} else {
			ch.Freeze(usr, e.Sub, opt, d)
		}

		raw[message.KEY_SUB] = e.Sub

	case message.CMD_EXTERN:
		if usr.Role < role.ASSISTANT {
			usr.Error(http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	data, err := json.Marshal(raw)
	if err != nil {
		usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch e.Mode {
	case message.MODE_UNI:
		sub, ok := ch.Find(e.Sub, usr.Group)
		if !ok {
			usr.Error(http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		sub.Send(data)
		usr.Send(data)

	case message.MODE_MULTI:
		g, ok := ch.Get(usr.Group)
		if !ok {
			usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		g.Foreach(func(id string, usr *user.User) {
			usr.Send(data)
		})

	case message.MODE_BROADCAST:
		ch.Foreach(func(id string, usr *user.User) {
			usr.Send(data)
		})

	default:
		usr.Error(http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (this *ChatHandler) control(usr *user.User, cid string, sub string, opt string, d time.Duration) (*CtrlResponse, error) {
	var (
		data CtrlResponse
		req  *http.Request
		res  *http.Response
	)

	cnf := this.conf[_ON_CONTROL].(utils.Conf)

	if cnf[key.ENABLE].(bool) {
		s, err := upstream.Get(cnf[key.UPSTREAM].(string))
		if err != nil {
			usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		req, err = http.NewRequest(cnf[key.METHOD].(string), s.URL(cnf, 80), nil)
		if err != nil {
			usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		dst := req.URL.Query()
		dst.Add("channel", cid)
		dst.Add("user", usr.ID)
		dst.Add("sub", sub)
		dst.Add("opt", opt)
		dst.Add("d", d.String())

		req.URL.RawQuery = dst.Encode()

		client := new(http.Client)

		res, err = client.Do(req)
		if err != nil {
			usr.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			usr.Error(http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return nil, err
		}
	} else {
		data.Status = http.StatusText(http.StatusOK)
		data.Code = http.StatusOK
	}

	return &data, nil
}

func (this *ChatHandler) onClose(e *Event.Event) {
	user := e.Target.(*user.User)

	ch, ok := channel.Get(user.Channel)
	if ok {
		ch.Remove(user)
	}
}

func (this *ChatHandler) getChannelId(path string) string {
	pat := this.conf[key.PATTERN].(string)
	tmp := []byte(path)[len(pat):]

	for i, c := range tmp {
		if c == '/' {
			return string(tmp[:i-1])
		}
	}

	return string(tmp)
}

func (this *ChatHandler) checkOrigin(r *http.Request) bool {
	if this.server.Regexp != nil {
		domain := r.Header.Get("Origin")
		if domain == "" {
			if r.Host == "" {
				return false
			}

			domain = r.Host
		}

		arr := this.server.Regexp.FindStringSubmatch(domain)
		if arr == nil {
			return false
		}
	}

	return true
}
