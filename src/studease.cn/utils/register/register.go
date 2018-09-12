package register

import (
	"reflect"
	"studease.cn/utils/log"
	"studease.cn/utils/log/level"
)

var (
	types map[string]reflect.Type
)

func init() {
	types = make(map[string]reflect.Type)
}

func Add(name string, typ reflect.Type) {
	types[name] = typ
	log.Debug(level.CORE, "add callback \"%s\"", name)
}

func Get(name string) (reflect.Type, bool) {
	typ, ok := types[name]
	return typ, ok
}
