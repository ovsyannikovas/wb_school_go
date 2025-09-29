package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Counter struct {
	Counter int32
}

type CounterAtomic struct {
	Counter atomic.Int32
}

type CounterMutex struct {
	Counter int32
	Mx      sync.Mutex
}

func (c *Counter) Increment() {
	c.Counter++
}

func (c *CounterAtomic) Increment() {
	c.Counter.Add(1)
}

func (c *CounterMutex) Increment() {
	c.Mx.Lock()
	defer c.Mx.Unlock()
	c.Counter++
}

func main() {
	atomicCounter := &CounterAtomic{}
	mutexCounter := &CounterMutex{}
	counter := &Counter{}

	const numGoroutines = 1000
	const incrementsPerGoroutine = 100

	var wg sync.WaitGroup

	// No sync
	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				counter.Increment()
			}
		}()
	}
	wg.Wait()

	// Atomic
	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				atomicCounter.Increment()
			}
		}()
	}
	wg.Wait()

	// Mutex
	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				mutexCounter.Increment()
			}
		}()
	}
	wg.Wait()

	fmt.Printf("No sync counter result: %d\n", counter.Counter)
	fmt.Printf("Atomic result: %d\n", atomicCounter.Counter.Load())
	fmt.Printf("Mutex result: %d\n", mutexCounter.Counter)
}
