package main

import (
	"context"
	"errors"
	"sync"
)

// ErrNoEvents возвращается, когда в очереди нет событий для публикации
var ErrNoEvents = errors.New("в outbox нет событий")

// Order представляет бизнес-сущность заказа
type Order struct {
	ID     int
	UserID int
}

// Event представляет событие, ожидающее публикации
type Event struct {
	ID      int
	Topic   string
	OrderID int
}

// Publisher определяет контракт для публикации событий во внешний брокер
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// Store — in-memory хранилище с поддержкой Transactional Outbox
type Store struct {
	mu      sync.Mutex
	relayMu sync.Mutex

	nextOrderID int
	nextEventID int

	orders map[int]Order
	outbox []Event
}

// NewStore создаёт и возвращает новое хранилище
func NewStore() *Store {
	return &Store{
		orders: make(map[int]Order),
	}
}

// CreateOrder атомарно создаёт заказ и добавляет событие в outbox
func (s *Store) CreateOrder(userID int) Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextOrderID++
	order := Order{
		ID:     s.nextOrderID,
		UserID: userID,
	}
	s.orders[order.ID] = order

	s.nextEventID++
	event := Event{
		ID:      s.nextEventID,
		Topic:   "orders.created",
		OrderID: order.ID,
	}
	s.outbox = append(s.outbox, event)

	return order
}

// PublishNext публикует самое старое событие из outbox и удаляет его при успехе
func (s *Store) PublishNext(ctx context.Context, publisher Publisher) error {
	s.relayMu.Lock()
	defer s.relayMu.Unlock()

	s.mu.Lock()
	if len(s.outbox) == 0 {
		s.mu.Unlock()
		return ErrNoEvents
	}

	event := s.outbox[0]
	s.mu.Unlock()

	if err := publisher.Publish(ctx, event); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.outbox = s.outbox[1:]
	return nil
}
