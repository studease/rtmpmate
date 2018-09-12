package http

import (
	"studease.cn/net/upstream"
	"studease.cn/utils"
	"studease.cn/utils/key"
	"studease.cn/utils/log"
)

var (
	include      = "conf/mime.types"
	default_type = "text/plain"
)

func Listen(name string, cnf utils.Conf) {
	var (
		v  interface{}
		ok bool
	)

	SERVER_NAME = name

	if v, ok = cnf[key.INCLUDE]; ok {
		include = v.(string)
	}

	if v, ok = cnf[key.DEFAULT_TYPE]; ok {
		default_type = v.(string)
	}

	if v, ok = cnf[key.UPSTREAM]; ok {
		upstream.Add(v.([]interface{}))
	}

	if v, ok = cnf[key.SERVER]; !ok {
		log.Error("http.Listen() failed: no server found")
		return
	}

	for _, e := range v.([]interface{}) {
		cnf := utils.Conf(e.(map[string]interface{}))
		s := new(Server).Init(cnf)

		go s.Listen()
	}
}
