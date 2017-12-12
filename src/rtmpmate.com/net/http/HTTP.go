package http

import (
	"net/http"
)

const (
	WEBROOT = "webroot"
)

var (
	COMMON_HEADERS http.Header
)

func init() {
	COMMON_HEADERS = make(http.Header)
	COMMON_HEADERS.Set("Server", "rtmpmate/0.0.30")
	COMMON_HEADERS.Set("Access-Control-Allow-Origin", "*")
}
