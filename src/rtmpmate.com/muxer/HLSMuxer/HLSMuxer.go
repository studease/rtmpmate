package HLSMuxer

import (
	"rtmpmate.com/muxer/FMP4Muxer"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"syscall"
)

type HLSMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New(dir string, name string) (*HLSMuxer, error) {
	var m HLSMuxer

	err := m.Init(dir, name, MuxerTypes.HLS)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *HLSMuxer) Init(dir string, name string, t string) error {
	err := this.FMP4Muxer.Init(dir, name, t)
	if err != nil {
		return err
	}

	return nil
}

func (this *HLSMuxer) GetM3U8() ([]byte, error) {
	return nil, syscall.EINVAL
}
