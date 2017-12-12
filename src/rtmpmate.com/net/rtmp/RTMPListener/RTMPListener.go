package RTMPListener

import (
	"fmt"
	"net"
	"os"
	MuxerTypes "rtmpmate.com/muxer/Types"
	RTMP "rtmpmate.com/net/rtmp"
	"rtmpmate.com/net/rtmp/Application"
	"rtmpmate.com/net/rtmp/Handshaker"
	"strconv"
)

type RTMPListener struct {
	tcpln   *net.TCPListener
	network string
	port    int
	exiting bool
}

func New() (*RTMPListener, error) {
	var ln RTMPListener
	return &ln, nil
}

func (this *RTMPListener) Listen(network string, port int) {
	if _, err := os.Stat(RTMP.APPLICATIONS); os.IsNotExist(err) {
		err = os.MkdirAll(RTMP.APPLICATIONS+"/", os.ModeDir)
		if err != nil {
			return
		}
	}

	if network == "" {
		network = "tcp4"
	}

	if port == 0 {
		port = 1935
	}
	address := strconv.Itoa(port)

	tcpaddr, err := net.ResolveTCPAddr(network, ":"+address)
	if err != nil {
		fmt.Printf("Failed to ResolveTCPAddr: %v.\n", err)
		return
	}

	tcpln, err := net.ListenTCP(network, tcpaddr)
	if err != nil {
		fmt.Printf("Failed to listen on port %d: %v.\n", port, err)
		return
	}
	this.tcpln = tcpln

	for this.exiting == false {
		conn, err := tcpln.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept RTMP connection: %v.\n", err)
			continue
		}

		go this.handler(conn)
	}

	fmt.Printf("%v exiting...\n", this)
}

func (this *RTMPListener) handler(conn net.Conn) {
	shaker, err := Handshaker.New(conn)
	if err != nil {
		fmt.Printf("Failed to create Handshaker: %v.\n", err)
		return
	}

	err = shaker.Shake()
	if err != nil {
		conn.Close()
		fmt.Printf("Failed to complete handshake: %v.\n", err)
		return
	}

	nc, err := Application.HandshakeComplete(conn)
	if err != nil {
		conn.Close()
		fmt.Printf("Failed to get NetConnection: %v.\n", err)
		return
	}

	nc.MuxerType = MuxerTypes.RTMP
	nc.Protocol = "rtmp"

	err = nc.Wait()
	if err != nil {
		fmt.Printf("Closing NetConnection: %v.\n", err)
	}

	nc.Close()
}
