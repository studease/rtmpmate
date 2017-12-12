package FLV

import (
	"bytes"
)

func GetFileHeader() ([]byte, error) {
	var b bytes.Buffer

	b.Write([]byte{
		'F', 'L', 'V',
		0x01,
		0x05,
		0x00, 0x00, 0x00, 0x09,
		0x00, 0x00, 0x00, 0x00,
	})

	return b.Bytes(), nil
}

func Format(tagType byte, size int, timestamp int, data []byte) ([]byte, error) {
	var b bytes.Buffer

	// tag header
	n := size
	t := timestamp
	b.Write([]byte{
		tagType,
		0xFF & byte(n>>16), 0xFF & byte(n>>8), 0xFF & byte(n),
		0xFF & byte(t>>24), 0xFF & byte(t>>16), 0xFF & byte(t>>8), 0xFF & byte(t),
		0x00, 0x00, 0x00,
	})

	// tag data
	b.Write(data)

	// tag size
	n += 11
	b.Write([]byte{
		0xFF & byte(n>>24), 0xFF & byte(n>>16), 0xFF & byte(n>>8), 0xFF & byte(n),
	})

	return b.Bytes(), nil
}
