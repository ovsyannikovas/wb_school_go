package sync

import (
	"sync/atomic"
)

type WaitGroup struct {
	counter atomic.Int64
	done    chan struct{}
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		done: make(chan struct{}),
	}
}

func (wg *WaitGroup) Add(delta int) {
	val := wg.counter.Add(int64(delta))
	if val < 0 {
		panic("negative WaitGroup counter")
	}
	if val == 0 {
		close(wg.done)
		wg.done = make(chan struct{})
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	<-wg.done
}
