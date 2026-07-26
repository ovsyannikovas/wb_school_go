package main

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	capacity   int           // макс. токенов
	rate       time.Duration // интервал пополнения одного токена
	tokens     int
	lastRefill time.Time
}

func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow возвращает true, если запрос разрешён.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Пополняем токены с момента последнего пополнения
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	newTokens := int(elapsed / tb.rate)
	if newTokens > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+newTokens)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}
