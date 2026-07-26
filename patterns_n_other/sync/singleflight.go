package sync

import (
	"sync"
)

// Подсказка:
// call 3 поля, sf 2 поля, Do() проверяем наличие, иначе создаем, вызываем, подтирвем

type call struct {
	val  interface{}
	err  error
	done chan struct{}
}

type SingleFlight struct {
	mu    sync.Mutex
	calls map[string]*call
}

func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		calls: make(map[string]*call),
	}
}

func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.Lock()

	if c, ok := sf.calls[key]; ok {
		sf.mu.Unlock()
		<-c.done
		return c.val, c.err
	}

	c := &call{done: make(chan struct{})}
	sf.calls[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn()
	close(c.done)

	sf.mu.Lock()
	delete(sf.calls, key)
	sf.mu.Unlock()

	return c.val, c.err
}
