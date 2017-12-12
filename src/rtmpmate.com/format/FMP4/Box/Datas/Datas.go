package Datas

import ()

var (
	FTYP = []byte{
		0x69, 0x73, 0x6F, 0x6D, // major_brand: isom
		0x0, 0x0, 0x0, 0x1, // minor_version: 0x01
		0x69, 0x73, 0x6F, 0x6D, // isom
		0x61, 0x76, 0x63, 0x31, // avc1
	}

	STSD_PREFIX = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x01, // entry_count
	}

	STTS = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // entry_count
	}

	STSC = STTS
	STCO = STTS

	STSZ = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // sample_size
		0x00, 0x00, 0x00, 0x00, // sample_count
	}

	HDLR_VIDEO = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // pre_defined
		0x76, 0x69, 0x64, 0x65, // handler_type: 'vide'
		0x00, 0x00, 0x00, 0x00, // reserved: 3 * 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x56, 0x69, 0x64, 0x65,
		0x6F, 0x48, 0x61, 0x6E,
		0x64, 0x6C, 0x65, 0x72, 0x00, // name: VideoHandler
	}

	HDLR_AUDIO = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // pre_defined
		0x73, 0x6F, 0x75, 0x6E, // handler_type: 'soun'
		0x00, 0x00, 0x00, 0x00, // reserved: 3 * 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x53, 0x6F, 0x75, 0x6E,
		0x64, 0x48, 0x61, 0x6E,
		0x64, 0x6C, 0x65, 0x72, 0x00, // name: SoundHandler
	}

	DREF = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x01, // entry_count
		0x00, 0x00, 0x00, 0x0C, // entry_size
		0x75, 0x72, 0x6C, 0x20, // type 'url '
		0x00, 0x00, 0x00, 0x01, // version(0) + flags
	}

	// Sound media header
	SMHD = []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // balance(2) + reserved(2)
	}

	// video media header
	VMHD = []byte{
		0x00, 0x00, 0x00, 0x01, // version(0) + flags
		0x00, 0x00, // graphicsmode: 2 bytes
		0x00, 0x00, 0x00, 0x00, // opcolor: 3 * 2 bytes
		0x00, 0x00,
	}
)

func GetSilentFrame(channelCount byte) []byte {
	var data []byte

	switch channelCount {
	case 0x01:
		data = []byte{
			0x00, 0xC8, 0x00, 0x80, 0x23, 0x80,
		}

	case 0x02:
		data = []byte{
			0x21, 0x00, 0x49, 0x90, 0x02, 0x19, 0x00, 0x23, 0x80,
		}

	case 0x03:
		data = []byte{
			0x00, 0xC8, 0x00, 0x80, 0x20, 0x84, 0x01, 0x26, 0x40, 0x08, 0x64,
			0x00, 0x8E,
		}

	case 0x04:
		data = []byte{
			0x00, 0xC8, 0x00, 0x80, 0x20, 0x84, 0x01, 0x26, 0x40, 0x08, 0x64,
			0x00, 0x80, 0x2C, 0x80, 0x08, 0x02, 0x38,
		}

	case 0x05:
		data = []byte{
			0x00, 0xC8, 0x00, 0x80, 0x20, 0x84, 0x01, 0x26, 0x40, 0x08, 0x64,
			0x00, 0x82, 0x30, 0x04, 0x99, 0x00, 0x21, 0x90, 0x02, 0x38,
		}

	case 0x06:
		data = []byte{
			0x00, 0xC8, 0x00, 0x80, 0x20, 0x84, 0x01, 0x26, 0x40, 0x08, 0x64,
			0x00, 0x82, 0x30, 0x04, 0x99, 0x00, 0x21, 0x90, 0x02, 0x00, 0xB2,
			0x00, 0x20, 0x08, 0xE0,
		}

	default:
	}

	return data
}
