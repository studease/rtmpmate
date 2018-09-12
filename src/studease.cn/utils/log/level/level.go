package level

import ()

const (
	INFO   = 1
	NOTICE = 2
	WARN   = 3
	ALERT  = 4
	ERROR  = 5
)

const (
	EVENT = 0x0010
	CORE  = 0x0020
	RTMP  = 0x0040
	HTTP  = 0x0080
	CHAT  = 0x0100
	ALL   = 0xFFF0
)
