package sync

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	locked atomic.Bool
}

const (
	unlocked = false
	locked   = true
)

func (s *SpinLock) Lock() {
	for !s.locked.CompareAndSwap(unlocked, locked) {
		runtime.Gosched()
	}
}

func (s *SpinLock) Unlock() {
	s.locked.Store(unlocked)
}
