package sync

import (
	"sync"
	"sync/atomic"
)

// Подсказка:
// 3 поля, одно атомик!, Add ==0, закрываем селектом, Done, Wait

type WaitGroup struct {
	counter atomic.Int64
	done    chan struct{}
	mu      sync.Mutex
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
		wg.mu.Lock()
		select {
		case <-wg.done:
		default:
			close(wg.done)
		}
		wg.mu.Unlock()
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	<-wg.done
}
