package AMF

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"math"
	"rtmpmate.com/net/rtmp/AMF/Types"
	"syscall"
)

type AMFHash struct {
	Hash map[string]*AMFValue
}

func (this *AMFHash) Init() {
	this.Hash = make(map[string]*AMFValue)
}

func (this *AMFHash) Get(key string) (*AMFValue, error) {
	return this.Hash[key], nil
}

func (this *AMFHash) ToString(depth int) string {
	var b bytes.Buffer
	b.WriteString("{")

	for k, v := range this.Hash {
		b.WriteString("\n" + string(bytes.Repeat([]byte("\t"), depth+1)) + k + ": ")

		switch v.Type {
		case Types.DOUBLE:
			fallthrough
		case Types.BOOLEAN:
			fallthrough
		case Types.STRING:
			fallthrough
		case Types.LONG_STRING:
			fmt.Fprintf(&b, "%v", v.Data)

		case Types.OBJECT:
			fallthrough
		case Types.ECMA_ARRAY:
			s := this.ToString(depth + 1)
			b.WriteString(s)

		case Types.STRICT_ARRAY:
			l := v.Data.(list.List)
			b.WriteString("[")
			for e := l.Front(); e != nil; e = e.Next() {
				fmt.Fprintf(&b, "%v", e.Value)

				if e.Next() != nil {
					b.WriteString(", ")
				}
			}
			b.WriteString("]")

		case Types.NULL:
			b.WriteString("null")

		case Types.UNDEFINED:
			b.WriteString("undefined")

		case Types.DATE:
			fmt.Fprintf(&b, "%v", v.Data)

		default:
		}
	}

	b.WriteString("\n" + string(bytes.Repeat([]byte("\t"), depth)) + "}")

	return string(b.Bytes())
}

type AMFValue struct {
	AMFHash
	Type       byte
	Key        string
	Data       interface{}
	Timtstamp  float64
	Timeoffset uint16
	Cost       int
	Ended      bool
}

type AMFObject struct {
	AMFHash
	Data  list.List
	Cost  int
	Ended bool
}

type AMFString struct {
	Data string
	Cost int
}

type AMFLongString struct {
	Data string
	Cost int
}

type AMFDate struct {
	Data       float64
	Timtstamp  float64
	Timeoffset uint16
	Cost       int
}

func Decode(data []byte, offset int, size int) (*AMFValue, error) {
	key, err := DecodeValue(data, offset, size)
	if err != nil {
		return nil, err
	}

	val, err := DecodeValue(data, offset+key.Cost, size-key.Cost)
	if err != nil {
		return nil, err
	}

	var v AMFValue
	v.Init()
	v.Type = val.Type
	v.Key = key.Data.(string)
	v.Data = val.Data
	v.Cost = key.Cost + val.Cost

	return &v, err
}

func DecodeString(data []byte, offset int, size int) (*AMFString, error) {
	if size < 2 {
		return nil, fmt.Errorf("Data not enough while decoding AMF string")
	}

	var v AMFString

	var pos = 0
	var length = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])
	pos += 2

	if length > 0 {
		v.Data = string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	v.Cost = pos

	return &v, nil
}

func DecodeObject(data []byte, offset int, size int) (*AMFObject, error) {
	if size < 3 {
		return nil, fmt.Errorf("Data not enough while decoding AMF object")
	}

	var v AMFObject
	v.Init()

	var pos = 0
	for v.Ended == false && pos < size {
		key, err := DecodeString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += key.Cost

		val, err := DecodeValue(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += val.Cost
		v.Ended = val.Ended

		val.Key = key.Data
		v.Data.PushBack(val)
		v.Hash[val.Key] = val
	}

	v.Cost = pos

	return &v, nil
}

func DecodeECMAArray(data []byte, offset int, size int) (*AMFObject, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF ECMA array")
	}

	var v AMFObject
	v.Init()

	var pos = 0
	var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])
	pos += 4

	for i := uint32(0); i < length; i++ {
		key, err := DecodeString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += key.Cost

		val, err := DecodeValue(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += val.Cost
		if i == length-1 {
			v.Ended = true
		}

		val.Key = key.Data
		v.Data.PushBack(val)
		v.Hash[val.Key] = val
	}

	v.Cost = pos

	return &v, nil
}

func DecodeStrictArray(data []byte, offset int, size int) (*AMFObject, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF ECMA array")
	}

	var v AMFObject
	v.Init()

	var pos = 0
	var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])
	pos += 4

	for i := uint32(0); i < length; i++ {
		val, err := DecodeValue(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += val.Cost
		if i == length-1 {
			v.Ended = true
		}

		v.Data.PushBack(val)
	}

	v.Cost = pos

	return &v, nil
}

func DecodeDate(data []byte, offset int, size int) (*AMFDate, error) {
	if size < 10 {
		return nil, fmt.Errorf("Data not enough while decoding AMF date")
	}

	var v AMFDate

	var pos = 0
	var bits = binary.BigEndian.Uint64(data[offset+pos : offset+pos+8])
	v.Timtstamp = math.Float64frombits(bits)
	pos += 8

	v.Timeoffset = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])
	v.Data = v.Timtstamp + float64(v.Timeoffset)*60*1000
	pos += 2

	v.Cost = pos

	return &v, nil
}

func DecodeLongString(data []byte, offset int, size int) (*AMFLongString, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF long string")
	}

	var v AMFLongString

	var pos = 0
	var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])
	pos += 4

	if length > 0 {
		v.Data = string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	v.Cost = pos

	return &v, nil
}

func DecodeValue(data []byte, offset int, size int) (*AMFValue, error) {
	if size < 1 {
		return nil, fmt.Errorf("Data not enough while decoding AMF value")
	}

	var v AMFValue
	v.Init()

	var pos = 0
	v.Type = data[offset+pos : offset+pos+1][0]
	pos += 1

	switch v.Type {
	case Types.DOUBLE:
		var bits = binary.BigEndian.Uint64(data[offset+pos : offset+pos+8])
		v.Data = math.Float64frombits(bits)
		pos += 8

	case Types.BOOLEAN:
		var bit = uint8(data[offset+pos : offset+pos+1][0])
		if bit > 0 {
			v.Data = true
		}
		pos += 1

	case Types.STRING:
		str, err := DecodeString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = str.Data
		pos += str.Cost

	case Types.OBJECT:
		obj, err := DecodeObject(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Hash = obj.Hash
		v.Data = obj.Data
		pos += obj.Cost

	case Types.NULL:
	case Types.UNDEFINED:

	case Types.ECMA_ARRAY:
		arr, err := DecodeECMAArray(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Hash = arr.Hash
		v.Data = arr.Data
		pos += arr.Cost

	case Types.END_OF_OBJECT:
		v.Ended = true

	case Types.STRICT_ARRAY:
		arr, err := DecodeStrictArray(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = arr.Data
		pos += arr.Cost

	case Types.DATE:
		date, err := DecodeDate(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = date.Data
		v.Timtstamp = date.Timtstamp
		v.Timeoffset = date.Timeoffset
		pos += date.Cost

	case Types.LONG_STRING:
		longstr, err := DecodeLongString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = longstr.Data
		pos += longstr.Cost

	default:
		pos = size
		return nil, fmt.Errorf("Skipping unsupported AMF value type(%x)", v.Type)
	}

	v.Cost = pos

	return &v, nil
}

type Encoder struct {
	buffer bytes.Buffer
}

func (this *Encoder) AppendBytes(b []byte) error {
	_, err := this.buffer.Write(b)
	return err
}

func (this *Encoder) AppendInt8(n int8) error {
	return this.buffer.WriteByte(byte(n))
}

func (this *Encoder) AppendInt16(n int16, littleEndian bool) error {
	var order binary.ByteOrder

	if littleEndian {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	err := binary.Write(&this.buffer, order, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) AppendInt32(n int32, littleEndian bool) error {
	var order binary.ByteOrder

	if littleEndian {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	err := binary.Write(&this.buffer, order, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) Encode() ([]byte, error) {
	return this.buffer.Bytes(), nil
}

func (this *Encoder) EncodeNumber(n float64) error {
	err := this.buffer.WriteByte(Types.DOUBLE)
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeBoolean(b bool) error {
	err := this.buffer.WriteByte(Types.BOOLEAN)
	if err != nil {
		return err
	}

	var tmp byte
	if b {
		tmp = 1
	} else {
		tmp = 0
	}

	err = this.buffer.WriteByte(tmp)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeString(s string) error {
	size := len(s)
	if size == 0 {
		return syscall.EINVAL
	}

	if size >= 0xFFFF {
		return this.encodeLongString(s)
	}

	err := this.buffer.WriteByte(Types.STRING)
	if err != nil {
		return err
	}

	tmp := uint16(size)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	_, err = this.buffer.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeObject(o *AMFObject) error {
	err := this.buffer.WriteByte(Types.OBJECT)
	if err != nil {
		return err
	}

	err = this.encodeProperties(o)
	if err != nil {
		return err
	}

	if o.Ended == false {
		return nil
	}

	tmp := uint16(0)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) encodeProperties(o *AMFObject) error {
	for item := o.Data.Front(); item != nil; item = item.Next() {
		v := item.Value.(*AMFValue)
		s := uint16(len(v.Key))
		if s > 0 {
			err := binary.Write(&this.buffer, binary.BigEndian, &s)
			if err != nil {
				return err
			}

			_, err = this.buffer.Write([]byte(v.Key))
			if err != nil {
				return err
			}
		}

		err := this.EncodeValue(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Encoder) EncodeNull() error {
	return this.buffer.WriteByte(byte(Types.NULL))
}

func (this *Encoder) EncodeUndefined() error {
	return this.buffer.WriteByte(byte(Types.UNDEFINED))
}

func (this *Encoder) EncodeECMAArray(o *AMFObject) error {
	err := this.buffer.WriteByte(Types.ECMA_ARRAY)
	if err != nil {
		return err
	}

	tmp := uint32(o.Data.Len())
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	this.encodeProperties(o)

	/*tmp2 := uint16(0)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp2)
	if err != nil {
		return err
	}

	err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}*/

	return nil
}

func (this *Encoder) EncodeStrictArray(o *AMFObject) error {
	err := this.buffer.WriteByte(Types.STRICT_ARRAY)
	if err != nil {
		return err
	}

	tmp := uint32(o.Data.Len())
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	this.encodeProperties(o)

	/*err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}*/

	return nil
}

func (this *Encoder) EncodeDate(ts float64, off uint16) error {
	err := this.buffer.WriteByte(byte(Types.DATE))
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &ts)
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &off)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) encodeLongString(ls string) error {
	size := len(ls)
	if size == 0 {
		return syscall.EINVAL
	}

	err := this.buffer.WriteByte(Types.LONG_STRING)
	if err != nil {
		return err
	}

	tmp := uint32(size)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	_, err = this.buffer.Write([]byte(ls))
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeValue(v *AMFValue) error {
	switch v.Type {
	case Types.DOUBLE:
		err := this.EncodeNumber(v.Data.(float64))
		if err != nil {
			return err
		}

	case Types.BOOLEAN:
		err := this.EncodeBoolean(v.Data.(bool))
		if err != nil {
			return err
		}

	case Types.STRING:
		err := this.EncodeString(v.Data.(string))
		if err != nil {
			return err
		}

	case Types.OBJECT:
		obj := AMFObject{AMFHash{v.Hash}, v.Data.(list.List), 0, true}
		err := this.EncodeObject(&obj)
		if err != nil {
			return err
		}

	case Types.NULL:
		err := this.EncodeNull()
		if err != nil {
			return err
		}

	case Types.UNDEFINED:
		err := this.EncodeUndefined()
		if err != nil {
			return err
		}

	case Types.ECMA_ARRAY:
		obj := AMFObject{AMFHash{v.Hash}, v.Data.(list.List), 0, true}
		err := this.EncodeECMAArray(&obj)
		if err != nil {
			return err
		}

	case Types.STRICT_ARRAY:
		obj := AMFObject{AMFHash{v.Hash}, v.Data.(list.List), 0, true}
		err := this.EncodeStrictArray(&obj)
		if err != nil {
			return err
		}

	case Types.DATE:
		err := this.EncodeDate(v.Timtstamp, v.Timeoffset)
		if err != nil {
			return err
		}

	case Types.LONG_STRING:
		err := this.encodeLongString(v.Data.(string))
		if err != nil {
			return err
		}

	default:
		fmt.Errorf("Skipping unsupported AMF value type(%x)", v.Type)
	}

	return nil
}

func (this *Encoder) Len() int {
	return this.buffer.Len()
}

func (this *Encoder) Reset() {
	this.buffer.Reset()
}
