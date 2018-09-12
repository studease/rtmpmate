package main

import (
	"runtime"
	"studease.cn/events/ErrorEvent"
	"studease.cn/events/Event"
	"studease.cn/events/SignalEvent"
	_ "studease.cn/net/chat"
	"studease.cn/net/http"
	"studease.cn/utils"
	"studease.cn/utils/key"
	"studease.cn/utils/license"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"studease.cn/utils/signal"
)

const (
	_NAME    = "rtmpmate"
	_VERSION = "0.0.31"
	_CONF    = "conf/mated.conf"
)

func main() {
	cnf := make(utils.Conf)
	cnf.Open(_CONF)

	log.New(cnf[key.LOG].(string), int(cnf[key.DEBUG].(float64)))
	log.Info("==== %s/%s ====", _NAME, _VERSION)
	log.Info("runtime: %s %s", runtime.GOOS, runtime.Version())

	lic, _ := license.New()
	lic.AddEventListener(Event.COMPLETE, onLicenseComplete, 0)
	lic.AddEventListener(ErrorEvent.ERROR, onLicenseError, 0)
	lic.Check(cnf[key.LICENSE].(string))

	http.Listen(_NAME+"/"+_VERSION, utils.Conf(cnf[key.HTTP].(map[string]interface{})))

	sig := signal.Default()
	sig.AddEventListener(SignalEvent.SIGNAL, onSignal, 0)
	sig.Wait()
}

func onLicenseComplete(e *Event.Event) {
	log.Debug(level.INFO, "license check complete")
}

func onLicenseError(e *ErrorEvent.ErrorEvent) {
	log.Fatal("license check error")
}

func onSignal(e *SignalEvent.SignalEvent) {
	switch e.Code {
	case signal.EXIT:
		log.Fatal("signal %d", e.Code)
	}
}
