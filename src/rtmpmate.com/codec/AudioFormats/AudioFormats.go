package AudioFormats

import ()

const (
	LINEAR_PCM_PLATFORM_ENDIAN   = 0x00
	ADPCM                        = 0x01
	MP3                          = 0x02
	LINEAR_PCM_LITTLE_ENDIAN     = 0x03
	NELLYMOSER_16_kHz_MONO       = 0x04
	NELLYMOSER_8_kHz_MONO        = 0x05
	NELLYMOSER                   = 0x06
	G_711_A_LAW_LOGARITHMIC_PCM  = 0x07
	G_711_MU_LAW_LOGARITHMIC_PCM = 0x08
	RESERVED                     = 0x09
	AAC                          = 0x0A
	SPEEX                        = 0x0B
	MP3_8_kHz                    = 0x0E
	DEVICE_SPECIFIC_SOUND        = 0x0F
)
