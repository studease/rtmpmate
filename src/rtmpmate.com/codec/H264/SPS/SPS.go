package SPS

import ()

type SPS struct {
	ProfileIdc         int
	LevelIdc           int
	ChromaFormatIdc    int
	BitDepth           int
	FixedFrameRateFlag int
	FPS                float64
	SarRatioWidth      int
	SarRatioHeight     int
	CodecWidth         int
	CodecHeight        int
	PresentWidth       int
	PresentHeight      int
}

func Parse(p []byte) (*SPS, error) {
	return nil, nil
}
