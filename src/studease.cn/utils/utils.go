package utils

import (
	"reflect"
)

func Class(typ reflect.Type) reflect.Value {
	return reflect.New(typ)
}
