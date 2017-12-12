package DASHMuxer

import (
	"os"
	"rtmpmate.com/muxer/FMP4Muxer"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"strconv"
)

const (
	MPD_FILENAME = "manifest"
)

type DASHMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New(dir string, name string) (*DASHMuxer, error) {
	var m DASHMuxer

	err := m.Init(dir, name, MuxerTypes.DASH)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *DASHMuxer) Init(dir string, name string, t string) error {
	name += "/m4s"
	path := dir + name + "/"
	os.RemoveAll(path)

	err := this.Muxer.Init(dir, name, t)
	if err != nil {
		return err
	}

	this.MaxBufferLength = 0x00200000
	this.MaxBufferTime = 3000
	this.LowLatency = true
	this.Record = true

	return nil
}

func (this *DASHMuxer) VideoHeaderGenerated(videoHeader []byte) {
	name := this.Dir + "video_video0_init_mp4.m4s"
	this.Save(name, videoHeader)
}

func (this *DASHMuxer) VideoSegmentGenerated(videoSegment []byte, timestamp int64, duration int) {
	name := this.Dir + "video_video0_" + strconv.Itoa(int(timestamp)) + "_mp4.m4s"
	this.Save(name, videoSegment)
}

func (this *DASHMuxer) AudioHeaderGenerated(audioHeader []byte) {
	name := this.Dir + "audio_audio0_init_mp4.m4s"
	this.Save(name, audioHeader)
}

func (this *DASHMuxer) AudioSegmentGenerated(audioSegment []byte, timestamp int64, duration int) {
	name := this.Dir + "audio_audio0_" + strconv.Itoa(int(timestamp)) + "_mp4.m4s"
	this.Save(name, audioSegment)
}

func (this *DASHMuxer) GetMPD() ([]byte, error) {
	return nil, nil
}
