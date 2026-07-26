package sync

import (
	"sync"
	"sync/atomic"
)

// Подсказка:
// 2 поля, Do(), одно атомик!

type Once struct {
	done atomic.Bool
	mu   sync.Mutex
}

func (o *Once) Do(f func()) {
	if o.done.Load() {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.done.Load() {
		return
	}

	f()
	o.done.Store(true)
}
