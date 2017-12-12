package Handshaker

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"rtmpmate.com/net/rtmp/Handshaker/Types"
	"syscall"
)

const (
	PACKET_SIZE = 1536
	DIGEST_SIZE = 32
	VERSION     = 0x5033029
)

var FP_KEY = []byte{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
	0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
	0x65, 0x72, 0x20, 0x30, 0x30, 0x31, /* Genuine Adobe Flash Player 001 */
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
	0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

var FMS_KEY = []byte{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
	0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x4D, 0x65, 0x64, 0x69,
	0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
	0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

type Handshaker struct {
	conn net.Conn
	mode uint8
}

func New(conn net.Conn) (*Handshaker, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	var shaker Handshaker
	shaker.conn = conn
	shaker.mode = Types.SIMPLE

	return &shaker, nil
}

func (this *Handshaker) Read(b []byte) (int, error) {
	size := len(b)

	for pos := 0; pos < size; {
		n, err := this.conn.Read(b[pos:])
		if err != nil {
			return pos, err
		}

		pos += n
	}

	return size, nil
}

func (this *Handshaker) Shake() error {
	var b = make([]byte, 1+PACKET_SIZE)
	_, err := this.Read(b)
	if err != nil {
		return err
	}

	if b[0] != 0x03 {
		return fmt.Errorf("Invalid handshake version: %d.\n", b[0])
	}

	zero := binary.BigEndian.Uint32(b[5:9])
	if zero == 0 {
		err = this.simpleHandshake(b[1:])
	} else {
		err = this.complexHandshake(b[1:])
	}

	/*if err != nil {
		return err
	}*/

	// Handshake done
	fmt.Printf("Handshake done.\n")

	return nil
}

func (this *Handshaker) simpleHandshake(c1 []byte) error {
	// S0 + S1
	s01 := make([]byte, 1+PACKET_SIZE)

	s01[0] = 0x03
	binary.BigEndian.PutUint32(s01[1:5], 0)
	binary.BigEndian.PutUint32(s01[5:9], 0)

	for i := 9; i <= PACKET_SIZE; i++ {
		s01[i] = byte(rand.Int() % 256)
	}

	n, err := this.conn.Write(s01)
	if err != nil {
		return err
	}

	if n != 1+PACKET_SIZE {
		return fmt.Errorf("Uncomplete to send S0 + S1: %d.\n", n)
	}

	// s2
	n, err = this.conn.Write(c1)
	if err != nil {
		return err
	}

	if n != PACKET_SIZE {
		return fmt.Errorf("Uncomplete to send S2: %d.\n", n)
	}

	// C2
	var c2 = make([]byte, PACKET_SIZE)
	_, err = this.Read(c2)
	if err != nil {
		return err
	}

	for i := 0; i < PACKET_SIZE; i++ {
		if c2[i] != s01[i+1] {
			return fmt.Errorf("Packet C2 & S1 not match.\n")
		}
	}

	return nil
}

func (this *Handshaker) complexHandshake(c1 []byte) error {
	var middle bool
	c1Digest, _, err := this.validateClient(c1, &middle)
	if err != nil {
		return fmt.Errorf("Invalid C1 data: %v.\n", err)
	}

	// S1
	s1 := this.getComplexS1()
	digestOffset, err := this.getDigestOffset(s1, middle)
	if err != nil {
		return fmt.Errorf("Failed to get S1 digest offset: %v.\n", err)
	}

	s1Random := make([]byte, PACKET_SIZE-DIGEST_SIZE)
	copy(s1Random, s1[:digestOffset])
	copy(s1Random[digestOffset:], s1[digestOffset+DIGEST_SIZE:])

	s1Hash := hmac.New(sha256.New, FMS_KEY[:36])
	s1Hash.Write(s1Random)

	s1Digest := s1Hash.Sum(nil)
	copy(s1[digestOffset:digestOffset+DIGEST_SIZE], s1Digest)

	// S2
	s2Hash := hmac.New(sha256.New, FMS_KEY[:68])
	s2Hash.Write(c1Digest)

	s2Tmp := this.getComplexS2()
	s2Digest := s2Hash.Sum(nil)

	s2Hash = hmac.New(sha256.New, s2Digest)
	s2Hash.Write(s2Tmp)

	s2 := s2Hash.Sum(nil)

	// send S0 S1 S2
	_, err = this.conn.Write([]byte{0x03})
	if err != nil {
		return err
	}

	_, err = this.conn.Write(s1)
	if err != nil {
		return err
	}

	_, err = this.conn.Write(s2Tmp)
	if err != nil {
		return err
	}

	_, err = this.conn.Write(s2)
	if err != nil {
		return err
	}

	// recv C2
	var c2 = make([]byte, PACKET_SIZE)
	_, err = this.Read(c2)
	if err != nil {
		return err
	}

	if len(c2) != PACKET_SIZE {
		return fmt.Errorf("Invalid C2")
	}

	for i := 0; i < PACKET_SIZE; i++ {
		if c2[i] != s1[i] {
			return fmt.Errorf("C2 != S1")
		}
	}

	return nil
}

func (this *Handshaker) validateClient(c1 []byte, middle *bool) ([]byte, []byte, error) {
	digest, challenge, err := this.validateClientScheme(c1, true)
	if err == nil {
		*middle = true
		return digest, challenge, nil
	}

	digest, challenge, err = this.validateClientScheme(c1, false)
	if err == nil {
		*middle = false
		return digest, challenge, nil
	}

	return nil, nil, fmt.Errorf("unknown scheme")
}

func (this *Handshaker) validateClientScheme(c1 []byte, middle bool) ([]byte, []byte, error) {
	digestOffset, err := this.getDigestOffset(c1, middle)
	if err != nil {
		return nil, nil, err
	}

	digest := make([]byte, DIGEST_SIZE)
	copy(digest, c1[digestOffset:digestOffset+DIGEST_SIZE])

	random := make([]byte, PACKET_SIZE-DIGEST_SIZE)
	copy(random, c1[:digestOffset])
	copy(random[digestOffset:], c1[digestOffset+DIGEST_SIZE:])

	hash := hmac.New(sha256.New, FP_KEY[:30])
	hash.Write(random)

	tmp := hash.Sum(nil)
	if bytes.Compare(tmp, digest) != 0 {
		return nil, nil, syscall.EINVAL
	}

	challengeOffset, err := this.getDHOffset(c1, middle)
	if err != nil {
		return nil, nil, err
	}

	challenge := make([]byte, 128)
	copy(challenge, c1[challengeOffset:challengeOffset+128])

	return digest, challenge, nil
}

func (this *Handshaker) getDigestOffset(c1 []byte, middle bool) (int, error) {
	offset := 8 + 4
	if middle {
		offset += 764
	}

	offset += (int(c1[offset-4]) + int(c1[offset-3]) + int(c1[offset-2]) + int(c1[offset-1])) % 728
	if offset+DIGEST_SIZE > PACKET_SIZE {
		return 0, fmt.Errorf("%d out of range", offset)
	}

	return offset, nil
}

func (this *Handshaker) getDHOffset(c1 []byte, middle bool) (int, error) {
	offset := 8 + 764
	if middle == false {
		offset += 764
	}

	offset = ((int(c1[offset-4]) + int(c1[offset-3]) + int(c1[offset-2]) + int(c1[offset-1])) % 632) + 8
	if middle == false {
		offset += 764
	}

	if offset+128 > PACKET_SIZE {
		return 0, fmt.Errorf("DH offset %d out of range", offset)
	}

	return offset, nil
}

func (this *Handshaker) getComplexS1() []byte {
	s1 := make([]byte, PACKET_SIZE)
	binary.BigEndian.PutUint32(s1[0:4], 0)
	binary.BigEndian.PutUint32(s1[4:8], VERSION)

	for i := 8; i < PACKET_SIZE; i++ {
		s1[i] = byte(rand.Int() % 256)
	}

	return s1
}

func (this *Handshaker) getComplexS2() []byte {
	size := PACKET_SIZE - DIGEST_SIZE
	s2 := make([]byte, size)

	for i := 0; i < size; i++ {
		s2[i] = byte(rand.Int() % 256)
	}

	return s2
}
