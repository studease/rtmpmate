package utils

import (
	"container/list"
	"sync"
)

type Queue struct {
	Data list.List
	sync.Mutex
}

func (this *Queue) Init() *Queue {
	this.Data.Init()
	return this
}
