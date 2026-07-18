package sync

import (
	"runtime"
	"sync/atomic"
)

type Mutex struct {
	state atomic.Bool
}

func (m *Mutex) Lock() {
	for !m.state.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
}

func (m *Mutex) Unlock() {
	m.state.Store(false)
}
