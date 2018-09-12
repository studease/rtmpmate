package http

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"studease.cn/utils"
	"studease.cn/utils/key"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"studease.cn/utils/register"
)

const (
	_PORT        = 80
	_TIMEOUT     = 65
	_SERVER_NAME = "^.+$"
	_ROOT        = "webroot"
	_CORS        = ""
)

var (
	SERVER_NAME = ""
)

type Handler interface {
	Init(*Server, utils.Conf)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type Server struct {
	Conf     utils.Conf
	Port     int
	Timeout  int
	Name     string
	Root     string
	Cors     string
	Location []interface{}
	Regexp   *regexp.Regexp
}

func (this *Server) Init(cnf utils.Conf) *Server {
	var (
		v   interface{}
		ok  bool
		err error
	)

	*this = Server{cnf, _PORT, _TIMEOUT, _SERVER_NAME, _ROOT, _CORS, nil, nil}

	if v, ok = cnf[key.LISTEN]; ok {
		this.Port = int(v.(float64))
	}

	if v, ok = cnf[key.TIMEOUT]; ok {
		this.Timeout = int(v.(float64))
	}

	if v, ok = cnf[key.SERVER_NAME]; ok {
		this.Name = v.(string)

		this.Regexp, err = regexp.Compile(this.Name)
		if err != nil {
			log.Warn("failed to compile regexp: %s", this.Name)
		}
	}

	if v, ok = cnf[key.ROOT]; ok {
		this.Root = v.(string)
	}

	if v, ok = cnf[key.CORS]; ok {
		this.Cors = v.(string)
	}

	if v, ok = cnf[key.LOCATION]; ok {
		this.Location = v.([]interface{})
	}

	return this
}

func (this *Server) Listen() {
	var (
		v       interface{}
		pattern string
		name    string
		typ     reflect.Type
		ok      bool
	)

	addr := ":" + strconv.Itoa(this.Port)

	for _, e := range this.Location {
		cnf := utils.Conf(e.(map[string]interface{}))

		if v, ok = cnf[key.PATTERN]; ok {
			pattern = v.(string)
		} else {
			pattern = "/"
		}

		if v, ok = cnf[key.HANDLER]; ok {
			name = v.(string)
		} else {
			name = FILE_HANDLER
		}

		typ, ok = register.Get(name)
		if !ok {
			log.Warn("unrecognized handler \"%s\"", name)
			continue
		}

		handler := utils.Class(typ).Interface().(Handler)
		handler.Init(this, cnf)

		http.HandleFunc(pattern, handler.ServeHTTP)
	}

	log.Debug(level.HTTP, "listening on port %d", this.Port)

	http.ListenAndServe(addr, nil)
}
