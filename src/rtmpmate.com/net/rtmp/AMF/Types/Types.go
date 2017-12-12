package Types

import ()

const (
	DOUBLE        = 0x00
	BOOLEAN       = 0x01
	STRING        = 0x02
	OBJECT        = 0x03
	MOVIE_CLIP    = 0x04 // Not available in Remoting
	NULL          = 0x05
	UNDEFINED     = 0x06
	REFERENCE     = 0x07
	ECMA_ARRAY    = 0x08
	END_OF_OBJECT = 0x09
	STRICT_ARRAY  = 0x0A
	DATE          = 0x0B
	LONG_STRING   = 0x0C
	UNSUPPORTED   = 0x0D
	RECORD_SET    = 0x0E // Remoting server-to-client only
	XML           = 0x0F
	TYPED_OBJECT  = 0x10 // Class instance
	AMF3_DATA     = 0x11 // Sent by Flash player 9+
)
