package HTTPListener

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"path"
	"regexp"
	"rtmpmate.com/muxer/DASHMuxer"
	HTTP "rtmpmate.com/net/http"
	RTMP "rtmpmate.com/net/rtmp"
	"rtmpmate.com/net/rtmp/Application"
	"strconv"
)

var (
	urlRe, _ = regexp.Compile("/([a-z0-9.-_]+)(/([a-z0-9.-_]+))?/([a-z0-9-_]+)/([a-z0-9-_]+)(.[a-z0-9]+)$")
)

type HTTPListener struct {
	network   string
	port      int
	websocket bool
	wsHandler func(w http.ResponseWriter, r *http.Request)
	exiting   bool
}

func New() (*HTTPListener, error) {
	var ln HTTPListener
	return &ln, nil
}

func (this *HTTPListener) Listen(network string, port int) {
	if _, err := os.Stat(HTTP.WEBROOT); os.IsNotExist(err) {
		err = os.MkdirAll(HTTP.WEBROOT+"/", os.ModeDir)
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

	http.HandleFunc("/", this.handler)
	http.ListenAndServe(":"+address, nil)

	fmt.Printf("%v exiting...\n", this)
}

func (this *HTTPListener) HandleWebSocket(handler func(http.ResponseWriter, *http.Request)) {
	this.websocket = true
	this.wsHandler = handler
}

func (this *HTTPListener) handler(w http.ResponseWriter, r *http.Request) {
	if this.websocket && websocket.IsWebSocketUpgrade(r) {
		this.wsHandler(w, r)
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if _, err := os.Stat(HTTP.WEBROOT + r.URL.Path); os.IsExist(err) {
		http.ServeFile(w, r, HTTP.WEBROOT+r.URL.Path)
		return
	}

	ext := path.Ext(r.URL.Path)
	switch ext {
	case ".mpd":
		this.serveMPD(w, r)
		return

	default:
	}

	if _, err := os.Stat(RTMP.APPLICATIONS + r.URL.Path); os.IsExist(err) {
		http.ServeFile(w, r, RTMP.APPLICATIONS+r.URL.Path)
		return
	}

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (this *HTTPListener) serveMPD(w http.ResponseWriter, r *http.Request) {
	arr := urlRe.FindStringSubmatch(r.URL.Path)
	if arr == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	appName := arr[1]
	instName := arr[3]
	streamName := arr[4]
	fileName := arr[5]
	if instName == "" {
		instName = "_definst_"
	}

	if fileName != DASHMuxer.MPD_FILENAME {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	app, err := Application.Get(appName)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	inst, _ := app.GetInstance(instName)
	stream, err := inst.GetStream(streamName, false)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	mpd, err := stream.DASHMuxer.GetMPD()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	n, err := w.Write(mpd)
	fmt.Printf("MPD: size=%d.\n", n)
}

func (this *HTTPListener) appendHeadersForStream(w http.ResponseWriter) {
	for k, v := range HTTP.COMMON_HEADERS {
		w.Header().Set(k, v[0])
	}
}
