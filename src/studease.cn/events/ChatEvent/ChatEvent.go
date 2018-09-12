package ChatEvent

import (
	"fmt"
	"studease.cn/events/Event"
)

const (
	MESSAGE = "ChatEvent.MESSAGE"
)

type ChatEvent struct {
	Event.Event
	Cmd  string
	Data string
	Mode int32
	Seq  int32
	Sub  string
}

func New(typ string, this interface{},
	cmd string, data string, mode int32, seq int32, sub string) *ChatEvent {
	return new(ChatEvent).Init(typ, this, cmd, data, mode, seq, sub)
}

func (e *ChatEvent) Init(typ string, this interface{},
	cmd string, data string, mode int32, seq int32, sub string) *ChatEvent {
	e.Event.Init(typ, this)
	e.Cmd = cmd
	e.Data = data
	e.Mode = mode
	e.Seq = seq
	e.Sub = sub
	return e
}

func (e *ChatEvent) Clone() *ChatEvent {
	return New(e.Type, e.Target, e.Cmd, e.Data, e.Mode, e.Seq, e.Sub)
}

func (e *ChatEvent) ToString() string {
	return fmt.Sprintf("[ChatEvent type=%s cmd=%s data=%s mode=%d seq=%d sub=%s]",
		e.Type, e.Cmd, e.Data, e.Mode, e.Seq, e.Sub)
}
