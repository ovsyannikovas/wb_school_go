package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// OutboxEvent — структура события
type OutboxEvent struct {
	ID          string
	AggregateID string
	EventType   string
	Payload     []byte
	Status      string // "PENDING", "SENT", "FAILED"
	Attempts    int
}

// InMemoryOutbox — хранилище событий в памяти
type InMemoryOutbox struct {
	mu     sync.Mutex
	events []OutboxEvent
	cond   *sync.Cond // для уведомления диспетчера о новых событиях
}

func NewInMemoryOutbox() *InMemoryOutbox {
	o := &InMemoryOutbox{}
	o.cond = sync.NewCond(&o.mu)
	return o
}

// AddEvent — атомарное добавление события (имитация транзакции)
func (o *InMemoryOutbox) AddEvent(event OutboxEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()
	event.Status = "PENDING"
	o.events = append(o.events, event)
	o.cond.Signal() // пробуждаем диспетчер
}

// DispatchLoop — фоновый процесс, отправляет события
func (o *InMemoryOutbox) DispatchLoop(ctx context.Context, sender func(OutboxEvent) error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("DispatchLoop stopped")
			return
		case <-ticker.C:
			o.processBatch(sender)
		}
	}
}

// processBatch — обрабатывает одну партию (до 100 событий)
func (o *InMemoryOutbox) processBatch(sender func(OutboxEvent) error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Находим все PENDING события (ограничим 100)
	var toProcess []int
	for i, ev := range o.events {
		if ev.Status == "PENDING" && len(toProcess) < 100 {
			toProcess = append(toProcess, i)
		}
	}
	if len(toProcess) == 0 {
		return
	}

	var sentIndices []int
	for _, idx := range toProcess {
		ev := &o.events[idx]
		err := sender(*ev)
		if err != nil {
			// Ретри с экспоненциальной задержкой (имитируем через attempts)
			ev.Attempts++
			if ev.Attempts >= 5 {
				ev.Status = "FAILED"
			}
			// В реальности здесь можно было бы вычислить available_at, но у нас нет времени, поэтому просто оставляем PENDING
		} else {
			ev.Status = "SENT"
			sentIndices = append(sentIndices, idx)
		}
	}
	// Можно также удалить SENT события, чтобы не занимали память, но для простоты оставим
}

// Пример бизнес-операции: создание пользователя
type UserService struct {
	outbox *InMemoryOutbox
	mu     sync.Mutex
	users  []string // имитация БД
}

func (s *UserService) CreateUser(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Сохраняем пользователя (в память)
	s.users = append(s.users, email)

	// 2. Генерируем событие
	event := OutboxEvent{
		ID:          fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		AggregateID: email, // упрощённо
		EventType:   "UserCreated",
		Payload:     []byte(`{"email":"` + email + `"}`),
	}
	// 3. Добавляем в outbox (атомарно с шагом 1 и 2, т.к. мы в одной критической секции)
	s.outbox.AddEvent(event)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outbox := NewInMemoryOutbox()

	// Запускаем диспетчер
	go outbox.DispatchLoop(ctx, func(ev OutboxEvent) error {
		log.Printf("[SENDER] sending event ID=%s type=%s payload=%s",
			ev.ID, ev.EventType, string(ev.Payload))
		// Имитация успеха (можно иногда возвращать ошибку для теста ретраев)
		return nil
	})

	// Создаём сервис пользователей
	userSvc := &UserService{outbox: outbox}

	// Создаём пользователя
	userSvc.CreateUser("alice@example.com")
	log.Println("User created")

	// Даём время на отправку
	time.Sleep(2 * time.Second)

	// Завершаем
	cancel()
	time.Sleep(200 * time.Millisecond)
}
