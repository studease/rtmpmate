package WebSocketListener

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"rtmpmate.com/net/rtmp/Application"
	WEBSOCKET "rtmpmate.com/net/websocket"
	"strconv"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
	urlRe, _ = regexp.Compile("/([a-z0-9.-_]+)(/([a-z0-9.-_]+))?/([a-z0-9.-_]+)$")
)

func init() {
	upgrader.CheckOrigin = checkOrigin
}

type WebSocketListener struct {
	network string
	port    int
	exiting bool
}

func New() (*WebSocketListener, error) {
	var ln WebSocketListener
	return &ln, nil
}

func (this *WebSocketListener) Listen(network string, port int) {
	if _, err := os.Stat(WEBSOCKET.WEBROOT); os.IsNotExist(err) {
		err = os.MkdirAll(WEBSOCKET.WEBROOT+"/", os.ModeDir)
		if err != nil {
			return
		}
	}

	if network == "" {
		network = "tcp4"
	}

	if port == 0 {
		port = 80
	}
	address := strconv.Itoa(port)

	http.HandleFunc("/", this.Handler)
	http.ListenAndServe(":"+address, nil)

	fmt.Printf("%v exiting...\n", this)
}

func (this *WebSocketListener) Handler(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) == false {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	conn, err := upgrader.Upgrade(w, r, WEBSOCKET.COMMON_HEADERS)
	if err != nil {
		fmt.Printf("Failed to upgrade to websocket: %v.\n", err)
		return
	}

	if _, err = os.Stat(WEBSOCKET.WEBROOT + r.URL.Path); os.IsExist(err) {
		b, err := ioutil.ReadFile(WEBSOCKET.WEBROOT + r.URL.Path)
		if err != nil {
			conn.Close()
			return
		}

		// TODO: write in frames
		conn.WriteMessage(websocket.BinaryMessage, b)

		return
	}

	nc, err := Application.HandshakeComplete(conn.UnderlyingConn())
	if err != nil {
		conn.Close()
		fmt.Printf("Failed to get NetConnection: %v.\n", err)
		return
	}

	nc.MuxerType = MuxerTypes.FMP4
	nc.Protocol = "ws"

	err = nc.WaitWebsocket(conn)
	if err != nil {
		fmt.Printf("Closing NetConnection: %v.\n", err)
	}

	nc.Close()
}

func checkOrigin(r *http.Request) bool {
	return true
}
