package main

import (
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	leakyBucketCh chan struct{}
	closeCh       chan struct{}
	closeDoneCh   chan struct{}
	mu            sync.Mutex
	isClosed      bool
}

func NewLeakyBucketLimiter(limit int, period time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		leakyBucketCh: make(chan struct{}, limit),
		closeCh:       make(chan struct{}),
		closeDoneCh:   make(chan struct{}),
	}

	// Заполняем ведро
	for i := 0; i < limit; i++ {
		limiter.leakyBucketCh <- struct{}{}
	}

	// Вычисляем интервал утечки
	leakInterval := period.Nanoseconds() / int64(limit)
	go limiter.startPeriodicLeak(time.Duration(leakInterval))

	return limiter
}

func (l *RateLimiter) startPeriodicLeak(interval time.Duration) {
	timer := time.NewTicker(interval)
	defer func() {
		timer.Stop()
		close(l.closeDoneCh)
	}()

	for {
		select {
		case <-l.closeCh:
			return
		case <-timer.C:
			// Удаляем одно разрешение из ведра
			select {
			case <-l.leakyBucketCh:
				// Разрешение удалено
			default:
				// Ведро пусто - ничего не делаем
			}
		}
	}
}

func (l *RateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isClosed {
		return false
	}

	select {
	case l.leakyBucketCh <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *RateLimiter) Shutdown() {
	l.mu.Lock()
	if l.isClosed {
		l.mu.Unlock()
		return
	}
	l.isClosed = true
	l.mu.Unlock()

	close(l.closeCh)
	<-l.closeDoneCh
	close(l.leakyBucketCh)
}

func main() {
	// Создаем лимитер: 5 запросов в секунду
	limiter := NewLeakyBucketLimiter(5, time.Second)
	defer limiter.Shutdown()

	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	// Запускаем 20 запросов
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if limiter.Allow() {
				mu.Lock()
				successCount++
				mu.Unlock()
				fmt.Printf("Запрос %d: РАЗРЕШЕН\n", id)
			} else {
				fmt.Printf("Запрос %d: ОТКЛОНЕН\n", id)
			}
		}(i)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	fmt.Printf("\nУспешных запросов: %d из 20\n", successCount)
}
