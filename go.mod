module github.com/studease/rtmpmate-source

go 1.12

replace (
	github.com/studease/common => ../common
	github.com/studease/common/av => ../common/av
	github.com/studease/common/av/codec => ../common/av/codec
	github.com/studease/common/av/format => ../common/av/format
	github.com/studease/common/av/utils => ../common/av/utils
	github.com/studease/common/chat => ../common/chat
	github.com/studease/common/dvr => ../common/dvr
	github.com/studease/common/events => ../common/events
	github.com/studease/common/http => ../common/http
	github.com/studease/common/log => ../common/log
	github.com/studease/common/mux => ../common/mux
	github.com/studease/common/target => ../common/target
	github.com/studease/common/utils => ../common/utils
	github.com/studease/common/utils/config => ../common/utils/config
	github.com/studease/common/utils/timer => ../common/utils/timer

	github.com/studease/rtmpmate/utils => ./src/utils
)

require (
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/studease/common/av v0.0.0-00010101000000-000000000000 // indirect
	github.com/studease/common/events v0.0.0-00010101000000-000000000000 // indirect
	github.com/studease/common/log v0.0.0-00010101000000-000000000000
	github.com/studease/common/mux v0.0.0-00010101000000-000000000000 // indirect
	github.com/studease/common/utils v0.0.0-00010101000000-000000000000 // indirect
	github.com/studease/rtmpmate-source/utils v0.0.0-00010101000000-000000000000
)
