package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CircuitBreakerState — состояние Circuit Breaker
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case Closed:
		return "CLOSED"
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker
type CircuitBreaker struct {
	state        CircuitBreakerState
	maxFailures  int
	resetTimeout time.Duration
	failureCount int
	lastFailure  time.Time
	mutex        sync.Mutex
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

// NewCircuitBreaker — создаём новый Circuit Breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Call — вызывает функцию с контролем состояния Circuit Breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	// Если breaker открыт — проверяем, не пора ли попробовать half-open
	if cb.state == Open {
		if time.Since(cb.lastFailure) >= cb.resetTimeout {
			cb.state = HalfOpen
		} else {
			cb.mutex.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mutex.Unlock()

	// Сам вызов делаем БЕЗ удержания мьютекса, чтобы не блокировать
	// чтение состояния на время сетевого запроса
	err := fn()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.state == HalfOpen {
			// Пробный запрос в half-open не удался — снова открываем
			cb.state = Open
			cb.failureCount = 0
		} else if cb.failureCount >= cb.maxFailures {
			cb.state = Open
		}
		return err
	}

	// Успех
	if cb.state == HalfOpen {
		cb.state = Closed
	}
	cb.failureCount = 0
	return nil
}

func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.state
}

// APIRequest — эмуляция вызова нестабильного API (60% ошибок)
func APIRequest(id int) error {
	time.Sleep(50 * time.Millisecond) // имитация сетевой задержки
	if rand.Intn(100) < 60 {
		return fmt.Errorf("API error for request %d", id)
	}
	return nil
}

func client(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	cb := NewCircuitBreaker(3, 5*time.Second)

	for i := 0; i < 10; i++ {
		reqID := id*100 + i // уникальный ID запроса для наглядности в логах
		for {
			err := cb.Call(func() error {
				return APIRequest(reqID)
			})

			if err == nil {
				fmt.Printf("[Client %d] request ID=%d -> SUCCESS (state=%s)\n", id, reqID, cb.State())
				break
			}

			if errors.Is(err, ErrCircuitOpen) {
				fmt.Printf("[Client %d] request ID=%d -> CIRCUIT OPEN, waiting for half-open...\n", id, reqID)
				time.Sleep(500 * time.Millisecond)
				continue // ждём и повторяем этот же запрос
			}

			// обычная ошибка API
			fmt.Printf("[Client %d] request ID=%d -> FAILURE: %v (state=%s)\n", id, reqID, err, cb.State())
			break
		}

		time.Sleep(300 * time.Millisecond)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go client(i, &wg)
	}
	wg.Wait()
	fmt.Println("All clients finished.")
}
