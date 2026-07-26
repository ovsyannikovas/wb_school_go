package main

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half-open"
)

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failCount    int
	successCount int           // для half-open
	threshold    int           // кол-во ошибок для перехода в open
	timeout      time.Duration // время в open перед переходом в half-open
	lastFailTime time.Time
	halfOpenMax  int // сколько пробных запросов пропускать в half-open (обычно 1)
}

func New(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		threshold:   threshold,
		timeout:     timeout,
		halfOpenMax: 1,
	}
}

// Call выполняет функцию fn, применяя логику Circuit Breaker.
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Проверяем, можно ли выполнить запрос
	if err := cb.allowRequest(); err != nil {
		return err
	}

	// Выполняем реальный вызов
	err := fn()

	// Обрабатываем результат
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		// Пропускаем только первые halfOpenMax запросов
		if cb.successCount < cb.halfOpenMax {
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Ошибка
		cb.failCount++
		cb.lastFailTime = time.Now()
		if cb.state == StateClosed && cb.failCount >= cb.threshold {
			cb.state = StateOpen
		}
		if cb.state == StateHalfOpen {
			// При ошибке в half-open снова уходим в open
			cb.state = StateOpen
			cb.failCount = 0
		}
		return
	}

	// Успех
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenMax {
			cb.state = StateClosed
			cb.failCount = 0
		}
		return
	}
	if cb.state == StateClosed {
		// Сбрасываем счётчик ошибок при успехе (можно сделать скользящее окно, но для простоты - сброс)
		cb.failCount = 0
	}
}
