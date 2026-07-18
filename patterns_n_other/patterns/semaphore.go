package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Semaphore struct {
	tokens chan struct{}
}

func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		tokens: make(chan struct{}, capacity),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.tokens <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) Release() {
	<-s.tokens
}

func main() {
	sem := NewSemaphore(3)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if err := sem.Acquire(context.Background()); err != nil {
				fmt.Printf("Горутина %d: не удалось получить семафор: %v\n", id, err)
				return
			}
			defer sem.Release()

			fmt.Printf("Горутина %d: работаю\n", id)
			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	fmt.Println("Все горутины завершились")
}
