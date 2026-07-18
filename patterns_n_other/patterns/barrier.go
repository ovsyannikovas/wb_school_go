package main

import (
	"fmt"
	"sync"
)

type Barrier interface {
	Await() error
	Reset()
}

type CyclicBarrier struct {
	mu      sync.Mutex
	count   int
	parties int
	ch      chan struct{}
}

func NewCyclicBarrier(parties int) *CyclicBarrier {
	return &CyclicBarrier{
		parties: parties,
		ch:      make(chan struct{}),
	}
}

func (b *CyclicBarrier) Await() error {
	b.mu.Lock()
	b.count++

	if b.count == b.parties {
		b.resetNoLock()
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	<-b.ch
	return nil
}

func (b *CyclicBarrier) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.resetNoLock()
}

func (b *CyclicBarrier) resetNoLock() {
	b.count = 0
	close(b.ch)
	b.ch = make(chan struct{})
}

// usage

func main() {
	barrier := NewCyclicBarrier(3)
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fmt.Printf("Горутина %d: ждет у барьера\n", id)
			barrier.Await()
			fmt.Printf("Горутина %d: прошла барьер!\n", id)

			// Барьер можно использовать повторно
			fmt.Printf("Горутина %d: снова ждет\n", id)
			barrier.Await()
			fmt.Printf("Горутина %d: снова прошла!\n", id)
		}(i)
	}

	wg.Wait()
}
