package main

import (
	"errors"
	"sync"
	"time"
)

var ErrBulkheadFull = errors.New("bulkhead limit reached")

type Bulkhead struct {
	sem chan struct{}
	wg  sync.WaitGroup // опционально для ожидания завершения всех задач
}

func NewBulkhead(limit int) *Bulkhead {
	return &Bulkhead{
		sem: make(chan struct{}, limit),
	}
}

// Execute выполняет функцию fn, если есть свободный слот.
// Если слотов нет – возвращает ErrBulkheadFull.
func (b *Bulkhead) Execute(fn func() error) error {
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn()
	default:
		return ErrBulkheadFull
	}
}

// ExecuteWithTimeout – вариант с таймаутом ожидания свободного слота
func (b *Bulkhead) ExecuteWithTimeout(fn func() error, timeout time.Duration) error {
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn()
	case <-time.After(timeout):
		return ErrBulkheadFull
	}
}
