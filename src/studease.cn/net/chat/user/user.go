package user

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"studease.cn/events"
	"studease.cn/events/ChatEvent"
	"studease.cn/events/Event"
	"studease.cn/net/chat/message"
	"studease.cn/utils"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"time"
)

type Info struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
	Role int32  `json:"role"`
}

type User struct {
	conn     *websocket.Conn
	cache    *utils.Queue
	active   time.Time
	interval time.Duration
	Channel  string
	Group    int32

	Info
	events.EventDispatcher
}

func New(conn *websocket.Conn, info *Info) *User {
	return new(User).Init(conn, info)
}

func (this *User) Init(conn *websocket.Conn, info *Info) *User {
	this.conn = conn
	this.cache = new(utils.Queue).Init()
	this.interval = getInterval(info.Role)
	this.Info = *info
	return this
}

func (this *User) Read() error {
	for {
		typ, b, err := this.conn.ReadMessage()
		if err != nil {
			return err
		}

		switch typ {
		case websocket.TextMessage, websocket.BinaryMessage:
			err = this.handler(b)
			if err != nil {
				return err
			}

		case websocket.CloseMessage:
			return errors.New("close")

		case websocket.PingMessage:
			log.Debug(level.INFO, "ping")

		case websocket.PongMessage:
			log.Debug(level.INFO, "pong")

		default:
			break
		}
	}

	return errors.New("unknown message type of websocket")
}

func (this *User) Send(data []byte) error {
	err := this.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Debug(level.CHAT, "failed to send data to user %s: %v", this.ID, err)
	}

	return err
}

func (this *User) SendRaw(raw map[string]interface{}) error {
	data, err := json.Marshal(raw)
	if err == nil {
		err = this.Send(data)
	}

	return err
}

func (this *User) Error(error string, code int) error {
	m := map[string]interface{}{
		message.KEY_CMD: message.CMD_ERROR,
		message.KEY_ERROR: map[string]interface{}{
			message.KEY_STATUS: error,
			message.KEY_CODE:   code,
		},
	}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return this.Send(b)
}

func (this *User) handler(data []byte) error {
	var (
		m message.Message
		//rights int32
	)

	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Debug(level.CHAT, "failed to unmarshal message")
		this.Error(http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return err
	}

	switch m.Cmd {
	case message.CMD_TEXT, message.CMD_CTRL, message.CMD_EXTERN:
		if this.Limited() {
			this.Error(http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		} else {
			this.DispatchEvent(ChatEvent.New(ChatEvent.MESSAGE, this, m.Cmd, m.Data, m.Mode, m.Seq, m.Sub))
		}

	case message.CMD_PING:
	case message.CMD_PONG:

	default:
		this.Error(http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return errors.New(http.StatusText(http.StatusMethodNotAllowed))
	}

	return nil
}

func (this *User) Limited() bool {
	now := time.Now()
	if now.Before(this.active.Add(this.interval)) {
		return true
	}

	this.active = now
	return false
}

func (this *User) Close() error {
	err := this.conn.Close()
	this.DispatchEvent(Event.New(Event.CLOSE, this))
	return err
}

func getInterval(role int32) time.Duration {
	if (role & 0xF0) != 0 {
		return 0
	}

	d := 2000

	if role != 0 {
		vip := role >> 1
		d = int(1000 - vip*100)
	}

	return time.Duration(d) * time.Millisecond
}
