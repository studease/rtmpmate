package http

import (
	"net/http"
	"reflect"
	"studease.cn/utils"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
	"studease.cn/utils/register"
)

const (
	FILE_HANDLER = "http_file"
)

func init() {
	register.Add(FILE_HANDLER, reflect.ValueOf(FileHandler{}).Type())
}

type FileHandler struct {
	server *Server
	conf   utils.Conf
}

func (this *FileHandler) Init(srv *Server, cnf utils.Conf) {
	this.server = srv
	this.conf = cnf
}

func (this *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if this.checkOrigin(r) == false {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	log.Debug(level.HTTP, "http file: %s", r.URL.Path)

	http.ServeFile(w, r, r.URL.Path)
}

func (this *FileHandler) checkOrigin(r *http.Request) bool {
	if this.server.Regexp != nil {
		domain := r.Header.Get("Origin")
		if domain == "" {
			if r.Host == "" {
				return false
			}

			domain = r.Host
		}

		arr := this.server.Regexp.FindStringSubmatch(domain)
		if arr == nil {
			return false
		}

		r.URL.Path = utils.ReplaceVar(this.server.Root, arr) + r.URL.Path
	}

	return true
}
