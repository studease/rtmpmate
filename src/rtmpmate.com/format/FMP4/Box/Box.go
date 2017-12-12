package Box

import (
	"bytes"
	"encoding/binary"
	"rtmpmate.com/format/FMP4/Box/Datas"
	"rtmpmate.com/format/FMP4/Box/Types"
	"rtmpmate.com/muxer/FMP4Muxer/Meta"
	"rtmpmate.com/muxer/FMP4Muxer/Track"
)

func Box(boxtype string, args ...[]byte) []byte {
	var box bytes.Buffer

	size := 8
	for _, v := range args {
		size += len(v)
	}

	binary.Write(&box, binary.BigEndian, &size) // set size
	box.WriteString(boxtype)                    // set type

	// set data
	for _, v := range args {
		box.Write(v)
	}

	return box.Bytes()
}

// Movie metadata box
func MOOV(meta *Meta.Meta) []byte {
	mvhd := MVHD(meta.Timescale, meta.Duration)
	trak := TRAK(meta)
	mvex := MVEX(meta)

	return Box(Types.MOOV, mvhd, trak, mvex)
}

// Movie header box
func MVHD(timescale uint32, duration uint32) []byte {
	return Box(Types.MVHD, []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // creation_time
		0x00, 0x00, 0x00, 0x00, // modification_time
		byte(timescale>>24) & 0xFF, // timescale: 4 bytes
		byte(timescale>>16) & 0xFF,
		byte(timescale>>8) & 0xFF,
		byte(timescale) & 0xFF,
		byte(duration>>24) & 0xFF, // duration: 4 bytes
		byte(duration>>16) & 0xFF,
		byte(duration>>8) & 0xFF,
		byte(duration) & 0xFF,
		0x00, 0x01, 0x00, 0x00, // Preferred rate: 1.0
		0x01, 0x00, 0x00, 0x00, // PreferredVolume(1.0, 2bytes) + reserved(2bytes)
		0x00, 0x00, 0x00, 0x00, // reserved: 4 + 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, // ----begin composition matrix----
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00, // ----end composition matrix----
		0x00, 0x00, 0x00, 0x00, // ----begin pre_defined 6 * 4 bytes----
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // ----end pre_defined 6 * 4 bytes----
		0xFF, 0xFF, 0xFF, 0xFF, // next_track_ID
	})
}

// Track box
func TRAK(meta *Meta.Meta) []byte {
	return Box(Types.TRAK, TKHD(meta), MDIA(meta))
}

// Track header box
func TKHD(meta *Meta.Meta) []byte {
	trackId := meta.ID
	duration := meta.Duration
	width := meta.PresentWidth
	height := meta.PresentHeight

	return Box(Types.TKHD, []byte{
		0x00, 0x00, 0x00, 0x07, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // creation_time
		0x00, 0x00, 0x00, 0x00, // modification_time
		byte(trackId>>24) & 0xFF, // track_ID: 4 bytes
		byte(trackId>>16) & 0xFF,
		byte(trackId>>8) & 0xFF,
		byte(trackId) & 0xFF,
		0x00, 0x00, 0x00, 0x00, // reserved: 4 bytes
		byte(duration>>24) & 0xFF, // duration: 4 bytes
		byte(duration>>16) & 0xFF,
		byte(duration>>8) & 0xFF,
		byte(duration) & 0xFF,
		0x00, 0x00, 0x00, 0x00, // reserved: 2 * 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // layer(2bytes) + alternate_group(2bytes)
		0x00, 0x00, 0x00, 0x00, // volume(2bytes) + reserved(2bytes)
		0x00, 0x01, 0x00, 0x00, // ----begin composition matrix----
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00, // ----end composition matrix----
		byte(width>>8) & 0xFF, // width and height
		byte(width) & 0xFF,
		0x00, 0x00,
		byte(height>>8) & 0xFF,
		byte(height) & 0xFF,
		0x00, 0x00,
	})
}

// Media Box
func MDIA(meta *Meta.Meta) []byte {
	return Box(Types.MDIA, MDHD(meta), HDLR(meta), MINF(meta))
}

// Media header box
func MDHD(meta *Meta.Meta) []byte {
	timescale := meta.Timescale
	duration := meta.Duration

	return Box(Types.MDHD, []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		0x00, 0x00, 0x00, 0x00, // creation_time
		0x00, 0x00, 0x00, 0x00, // modification_time
		byte(timescale>>24) & 0xFF, // timescale: 4 bytes
		byte(timescale>>16) & 0xFF,
		byte(timescale>>8) & 0xFF,
		byte(timescale) & 0xFF,
		byte(duration>>24) & 0xFF, // duration: 4 bytes
		byte(duration>>16) & 0xFF,
		byte(duration>>8) & 0xFF,
		byte(duration) & 0xFF,
		0x55, 0xC4, // language: und (undetermined)
		0x00, 0x00, // pre_defined = 0
	})
}

// Media handler reference box
func HDLR(meta *Meta.Meta) []byte {
	var data []byte

	if meta.Type == "audio" {
		data = Datas.HDLR_AUDIO
	} else {
		data = Datas.HDLR_VIDEO
	}

	return Box(Types.HDLR, data)
}

// Media infomation box
func MINF(meta *Meta.Meta) []byte {
	var xmhd []byte

	if meta.Type == "audio" {
		xmhd = Box(Types.SMHD, Datas.SMHD)
	} else {
		xmhd = Box(Types.VMHD, Datas.VMHD)
	}

	return Box(Types.MINF, xmhd, DINF(), STBL(meta))
}

// Data infomation box
func DINF() []byte {
	return Box(Types.DINF, Box(Types.DREF, Datas.DREF))
}

// Sample table box
func STBL(meta *Meta.Meta) []byte {
	var result = Box(Types.STBL, // type: stbl
		STSD(meta),                  // Sample Description Table
		Box(Types.STTS, Datas.STTS), // Time-To-Sample
		Box(Types.STSC, Datas.STSC), // Sample-To-Chunk
		Box(Types.STSZ, Datas.STSZ), // Sample size
		Box(Types.STCO, Datas.STCO), // Chunk offset
	)

	return result
}

// Sample description box
func STSD(meta *Meta.Meta) []byte {
	if meta.Type == "audio" {
		return Box(Types.STSD, Datas.STSD_PREFIX, MP4A(meta))
	} else {
		return Box(Types.STSD, Datas.STSD_PREFIX, AVC1(meta))
	}
}

func MP4A(meta *Meta.Meta) []byte {
	channelCount := meta.ChannelCount
	sampleRate := meta.SampleRate

	data := []byte{
		0x00, 0x00, 0x00, 0x00, // reserved(4)
		0x00, 0x00, 0x00, 0x01, // reserved(2) + data_reference_index(2)
		0x00, 0x00, 0x00, 0x00, // reserved: 2 * 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, channelCount, // channelCount(2)
		0x00, 0x10, // sampleSize(2)
		0x00, 0x00, 0x00, 0x00, // reserved(4)
		byte(sampleRate>>8) & 0xFF, // Audio sample rate
		byte(sampleRate) & 0xFF,
		0x00, 0x00,
	}

	return Box(Types.MP4A, data, ESDS(meta))
}

func ESDS(meta *Meta.Meta) []byte {
	var data bytes.Buffer

	config := meta.ChannelConfig
	configSize := byte(len(config))

	data.Write([]byte{
		0x00, 0x00, 0x00, 0x00, // version 0 + flags

		0x03,              // descriptor_type
		0x17 + configSize, // length3
		0x00, 0x01,        // es_id
		0x00, // stream_priority

		0x04,              // descriptor_type
		0x0F + configSize, // length
		0x40,              // codec: mpeg4_audio
		0x15,              // stream_type: Audio
		0x00, 0x00, 0x00,  // buffer_size
		0x00, 0x00, 0x00, 0x00, // maxBitrate
		0x00, 0x00, 0x00, 0x00, // avgBitrate

		0x05, // descriptor_type
	})
	data.WriteByte(configSize)
	data.Write(config)
	data.Write([]byte{
		0x06, 0x01, 0x02, // GASpecificConfig
	})

	return Box(Types.ESDS, data.Bytes())
}

func AVC1(meta *Meta.Meta) []byte {
	avcc := meta.AVCC
	width := meta.CodecWidth
	height := meta.CodecHeight

	data := []byte{
		0x00, 0x00, 0x00, 0x00, // reserved(4)
		0x00, 0x00, 0x00, 0x01, // reserved(2) + data_reference_index(2)
		0x00, 0x00, 0x00, 0x00, // pre_defined(2) + reserved(2)
		0x00, 0x00, 0x00, 0x00, // pre_defined: 3 * 4 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		byte(width>>8) & 0xFF, // width: 2 bytes
		byte(width) & 0xFF,
		byte(height>>8) & 0xFF, // height: 2 bytes
		byte(height) & 0xFF,
		0x00, 0x48, 0x00, 0x00, // horizresolution: 4 bytes
		0x00, 0x48, 0x00, 0x00, // vertresolution: 4 bytes
		0x00, 0x00, 0x00, 0x00, // reserved: 4 bytes
		0x00, 0x01, // frame_count
		0x0A,                   // strlen
		0x78, 0x71, 0x71, 0x2F, // compressorname: 32 bytes
		0x66, 0x6C, 0x76, 0x2E,
		0x6A, 0x73, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x18, // depth
		0xFF, 0xFF, // pre_defined = -1
	}

	return Box(Types.AVC1, data, Box(Types.AVCC, avcc))
}

// Movie Extends box
func MVEX(meta *Meta.Meta) []byte {
	return Box(Types.MVEX, TREX(meta))
}

// Track Extends box
func TREX(meta *Meta.Meta) []byte {
	trackId := meta.ID
	data := []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) + flags
		byte(trackId>>24) & 0xFF, // track_ID
		byte(trackId>>16) & 0xFF,
		byte(trackId>>8) & 0xFF,
		byte(trackId) & 0xFF,
		0x00, 0x00, 0x00, 0x01, // default_sample_description_index
		0x00, 0x00, 0x00, 0x00, // default_sample_duration
		0x00, 0x00, 0x00, 0x00, // default_sample_size
		0x00, 0x01, 0x00, 0x01, // default_sample_flags
	}

	return Box(Types.TREX, data)
}

// Movie fragment box
func MOOF(track *Track.Track, baseMediaDecodeTime uint32) []byte {
	return Box(Types.MOOF, MFHD(track.SequenceNumber), TRAF(track, baseMediaDecodeTime))
}

func MFHD(sequenceNumber uint32) []byte {
	data := []byte{
		0x00, 0x00, 0x00, 0x00,
		byte(sequenceNumber>>24) & 0xFF, // sequence_number: int32
		byte(sequenceNumber>>16) & 0xFF,
		byte(sequenceNumber>>8) & 0xFF,
		byte(sequenceNumber) & 0xFF,
	}

	return Box(Types.MFHD, data)
}

// Track fragment box
func TRAF(track *Track.Track, baseMediaDecodeTime uint32) []byte {
	trackId := track.ID

	// Track fragment header box
	tfhd := Box(Types.TFHD, []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) & flags
		byte(trackId>>24) & 0xFF, // track_ID
		byte(trackId>>16) & 0xFF,
		byte(trackId>>8) & 0xFF,
		byte(trackId) & 0xFF,
	})

	// Track Fragment Decode Time
	tfdt := Box(Types.TFDT, []byte{
		0x00, 0x00, 0x00, 0x00, // version(0) & flags
		byte(baseMediaDecodeTime>>24) & 0xFF, // baseMediaDecodeTime: int32
		byte(baseMediaDecodeTime>>16) & 0xFF,
		byte(baseMediaDecodeTime>>8) & 0xFF,
		byte(baseMediaDecodeTime) & 0xFF,
	})

	sdtp := SDTP(track)
	trun := TRUN(track, len(sdtp)+16+16+8+16+8+8)

	return Box(Types.TRAF, tfhd, tfdt, trun, sdtp)
}

// Sample Dependency Type box
func SDTP(track *Track.Track) []byte {
	samples := track.Samples
	data := make([]byte, 4+samples.Len())

	pos := 4

	// 0~4 bytes: version(0) & flags
	for e := samples.Front(); e != nil; e = e.Next() {
		sample := e.Value.(*Track.Sample)
		flags := sample.Flags
		data[pos] = (flags.IsLeading << 6) | // is_leading: 2 (bit)
			(flags.DependsOn << 4) | // sample_depends_on
			(flags.IsDependedOn << 2) | // sample_is_depended_on
			(flags.HasRedundancy) // sample_has_redundancy

		pos++
	}

	return Box(Types.SDTP, data)
}

// Track fragment run box
func TRUN(track *Track.Track, offset int) []byte {
	var data bytes.Buffer

	samples := track.Samples
	sampleCount := samples.Len()
	dataSize := 12 + 16*sampleCount

	offset += 8 + dataSize

	data.Write([]byte{
		0x00, 0x00, 0x0F, 0x01, // version(0) & flags
	})
	binary.Write(&data, binary.BigEndian, &sampleCount) // sample_count
	binary.Write(&data, binary.BigEndian, &offset)      // data_offset

	for e := samples.Front(); e != nil; e = e.Next() {
		sample := e.Value.(*Track.Sample)
		duration := sample.Duration
		size := sample.Size
		flags := sample.Flags
		cts := sample.CTS

		binary.Write(&data, binary.BigEndian, &duration) // sample_duration
		binary.Write(&data, binary.BigEndian, &size)     // sample_size
		data.Write([]byte{
			(flags.IsLeading << 2) | flags.DependsOn, // sample_flags
			(flags.IsDependedOn << 6) | (flags.HasRedundancy << 4) | flags.IsNonSync,
			0x00, 0x00, // sample_degradation_priority
		})
		binary.Write(&data, binary.BigEndian, &cts)
	}

	return Box(Types.TRUN, data.Bytes())
}

func MDAT(data []byte) []byte {
	return Box(Types.MDAT, data)
}
