package log

import (
	"log"
	"os"
	lv "studease.cn/utils/log/level"
	"time"
)

const (
	_DEBUG = "[DEBUG] "
	_INFO  = "[INFO ] "
	_WARN  = "[WARN ] "
	_ERROR = "[ERROR] "
	_FATAL = "[FATAL] "
)

const (
	DEFAULT_PATH = "logs/2006-01-02 15-04-05.log"
)

var (
	_log   *log.Logger = log.New(os.Stderr, "", log.LstdFlags)
	_debug int         = lv.ALL
)

func New(path string, debug int) {
	if path == "" {
		path = DEFAULT_PATH
	}

	if debug <= 0 {
		debug = lv.ALL
	}

	path = time.Now().Format(path)

	i := len(path)
	for i > 0 && !os.IsPathSeparator(path[i-1]) {
		i--
	}

	if i > 0 {
		err := os.MkdirAll(path[:i-1], os.ModePerm)
		if err != nil {
			Fatal("%s", err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		Fatal("%s", err)
	}

	_log = log.New(f, "", log.LstdFlags)
	_debug = debug
}

func Debug(level int, format string, v ...interface{}) {
	if level&_debug&lv.ALL != 0 || level >= _debug&0x0F {
		_log.SetPrefix(_DEBUG)
		_log.Printf(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	_log.SetPrefix(_INFO)
	_log.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
	_log.SetPrefix(_WARN)
	_log.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	_log.SetPrefix(_ERROR)
	_log.Printf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	_log.SetPrefix(_FATAL)
	_log.Fatalf(format, v...)
}
