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

// CircuitBreaker
type CircuitBreaker struct {
	state         CircuitBreakerState
	maxFailures   int
	resetTimeout  time.Duration
	failureCount  int
	lastFailure   time.Time
	mutex         sync.RWMutex
	halfOpenTimer *time.Timer
}

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
	for {
		cb.mutex.Lock()

		switch cb.state {
		case Open:
			// Проверяем, прошло ли достаточно времени
			if time.Since(cb.lastFailure) >= cb.resetTimeout {
				cb.state = HalfOpen
				cb.mutex.Unlock()
			} else {
				// Circuit ещё открыт — ждём
				wait := cb.resetTimeout - time.Since(cb.lastFailure)
				cb.mutex.Unlock()

				time.Sleep(wait)
				continue
			}

		case Closed:
			cb.mutex.Unlock()

		case HalfOpen:
			// В HalfOpen разрешаем один пробный запрос.
			cb.mutex.Unlock()
		}

		// Выполняем запрос
		err := fn()

		cb.mutex.Lock()
		defer cb.mutex.Unlock()

		if err == nil {
			// Успешный запрос:
			// в HalfOpen возвращаемся в Closed.
			cb.failureCount = 0
			cb.state = Closed

			return nil
		}

		// Ошибка запроса
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.state == HalfOpen {
			// Пробный запрос в HalfOpen снова завершился ошибкой.
			// Снова открываем circuit.
			cb.state = Open
			return err
		}

		if cb.failureCount >= cb.maxFailures {
			cb.state = Open
		}

		return err
	}
}

// APIRequest — эмуляция вызова API
func APIRequest(id int) error {
	// 60% вероятность ошибки
	if rand.Intn(100) < 60 {
		return errors.New("API request failed")
	}

	return nil
}

func client(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	cb := NewCircuitBreaker(3, 5*time.Second)

	for i := 0; i < 10; i++ {
		err := cb.Call(func() error {
			return APIRequest(i)
		})

		if err != nil {
			if err.Error() == "circuit breaker is open" {
				fmt.Printf(
					"Client %d: circuit breaker open for ID %d\n",
					id,
					i,
				)
			} else {
				fmt.Printf(
					"Client %d: request to ID %d failed: %v\n",
					id,
					i,
					err,
				)
			}
		} else {
			fmt.Printf(
				"Client %d: request to ID %d succeeded\n",
				id,
				i,
			)
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
