package Types

import ()

const (
	SET_CHUNK_SIZE     byte = 0x01
	ABORT              byte = 0x02
	ACK                byte = 0x03
	USER_CONTROL       byte = 0x04
	ACK_WINDOW_SIZE    byte = 0x05
	BANDWIDTH          byte = 0x06
	EDGE               byte = 0x07
	AUDIO              byte = 0x08
	VIDEO              byte = 0x09
	AMF3_DATA          byte = 0x0F
	AMF3_SHARED_OBJECT byte = 0x10
	AMF3_COMMAND       byte = 0x11
	DATA               byte = 0x12
	SHARED_OBJECT      byte = 0x13
	COMMAND            byte = 0x14
	AGGREGATE          byte = 0x16
)
