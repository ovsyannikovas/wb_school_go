package sync

import (
	"sync"
)

// Подсказка:
// 3 поля, waiters, Wait, Signal, Broadcast

type Cond struct {
	L       sync.Locker
	waiters []chan struct{}
	mu      sync.Mutex
}

func NewCond(l sync.Locker) *Cond {
	return &Cond{
		L: l,
	}
}

func (c *Cond) Wait() {
	ch := make(chan struct{})
	c.mu.Lock()
	c.waiters = append(c.waiters, ch)
	c.mu.Unlock()

	c.L.Unlock()
	<-ch
	c.L.Lock()
}

func (c *Cond) Signal() {
	c.mu.Lock()
	if len(c.waiters) == 0 {
		c.mu.Unlock()
		return
	}
	ch := c.waiters[0]
	c.waiters = c.waiters[1:]
	c.mu.Unlock()
	close(ch)
}

func (c *Cond) Broadcast() {
	c.mu.Lock()
	waiters := c.waiters
	c.waiters = nil
	c.mu.Unlock()

	for _, ch := range waiters {
		close(ch)
	}
}
