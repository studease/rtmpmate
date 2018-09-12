package message

import (
	"regexp"
	"strconv"
	"time"
)

const (
	KEY_CMD     = "cmd"
	KEY_SEQ     = "seq"
	KEY_DATA    = "data"
	KEY_MODE    = "mode"
	KEY_USER    = "user"
	KEY_SUB     = "sub"
	KEY_CHANNEL = "channel"
	KEY_GROUP   = "group"
	KEY_ERROR   = "error"

	KEY_ID     = "id"
	KEY_NAME   = "name"
	KEY_ICON   = "icon"
	KEY_ROLE   = "role"
	KEY_TOTAL  = "total"
	KEY_STAT   = "stat"
	KEY_STATUS = "status"
	KEY_CODE   = "code"
)

const (
	CMD_INFO   = "info"
	CMD_TEXT   = "text"
	CMD_USER   = "user"
	CMD_JOIN   = "join"
	CMD_LEFT   = "left"
	CMD_CTRL   = "ctrl"
	CMD_EXTERN = "extern"
	CMD_PING   = "ping"
	CMD_PONG   = "pong"
	CMD_ERROR  = "error"
)

const (
	OPT_MUTE   = "mute"
	OPT_FORBID = "forbid"
)

const (
	MODE_UNI       = 0x00
	MODE_MULTI     = 0x01
	MODE_BROADCAST = 0x02
	MODE_OUTDATED  = 0x04
)

var (
	optRe, _ = regexp.Compile("^([a-z]+):([0-9]+)$")
)

type Message struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
	Mode int32  `json:"mode"`
	Seq  int32  `json:"seq"`
	Sub  string `json:"sub"`
}

func New(cmd string, data string, mode int32, seq int32, sub string) *Message {
	return new(Message).Init(cmd, data, mode, seq, sub)
}

func (this *Message) Init(cmd string, data string, mode int32, seq int32, sub string) *Message {
	this.Cmd = cmd
	this.Data = data
	this.Mode = mode
	this.Seq = seq
	this.Sub = sub
	return this
}

func ParseControl(data string) (string, time.Duration) {
	arr := optRe.FindStringSubmatch(data)
	if arr == nil {
		return "", time.Duration(0)
	}

	n, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil {
		return "", time.Duration(0)
	}

	return arr[1], time.Duration(n) * time.Second
}
