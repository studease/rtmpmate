package utils

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"studease.cn/utils/log"
)

const (
	CONF_MAX_SIZE = 1048576
)

var (
	reg, _ = regexp.Compile("(\\$([0-9].))?")
)

type Conf map[string]interface{}

func (this Conf) Open(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("%s", err)
	}

	buf := make([]byte, CONF_MAX_SIZE)
	n := 0

	for {
		i, err := f.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal("%s", err)
		}

		n += i

		if i == 0 {
			break
		}
	}

	err = json.Unmarshal(buf[:n], &this)
	if err != nil {
		log.Fatal("%s", err)
	}
}

func Extends(dst Conf, arr ...Conf) Conf {
	var (
		cnf, tmp0, tmp1 interface{}
		ok              bool
	)

	for _, src := range arr {
		for k, v := range src {
			if reflect.ValueOf(v).Kind() == reflect.Map {
				cnf, ok = dst[k]
				if !ok {
					cnf = make(Conf)
				}

				tmp0, ok = cnf.(Conf)
				if !ok {
					tmp0 = Conf(cnf.(map[string]interface{}))
				}

				tmp1, ok = v.(Conf)
				if !ok {
					tmp1 = Conf(v.(map[string]interface{}))
				}

				dst[k] = Extends(tmp0.(Conf), tmp1.(Conf))
			} else {
				dst[k] = v
			}
		}
	}

	return dst
}

func ReplaceVar(src string, arr []string) string {
	const (
		sw_searching = iota
		sw_mark
		sw_value
	)

	dst := ""
	index := 0
	state := sw_searching

	for _, c := range []byte(src) {
		if state == sw_searching {
			if c == '$' {
				state = sw_mark
				continue
			}

			dst += string(c)
			continue
		}

		if c < '0' || c > '9' {
			if state == sw_value {
				if index < len(arr) {
					dst += arr[index]
				}
			}

			if c == '$' {
				state = sw_mark
				continue
			}

			dst += string(c)
			state = sw_searching
			continue
		}

		if state == sw_mark {
			index, _ = strconv.Atoi(string(c))
			state = sw_value
			continue
		}

		if state == sw_value {
			tmp, _ := strconv.Atoi(string(c))
			index *= 10
			index += tmp
			continue
		}

		log.Warn("unknown state: %d", state)
	}

	if state == sw_value && index < len(arr) {
		dst += arr[index]
	}

	return dst
}
