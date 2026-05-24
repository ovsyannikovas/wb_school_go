package memory

import (
	"context"
	"L4_3/internal/domain"
	"sync"
	"time"
)

type EventRepository struct {
	mu         sync.RWMutex
	events     map[string]*domain.Event
	userEvents map[string][]string // userID -> []eventID
	archived   []*domain.ArchivedEvent
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events:     make(map[string]*domain.Event),
		userEvents: make(map[string][]string),
		archived:   make([]*domain.ArchivedEvent, 0),
	}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	r.events[event.ID] = event
	r.userEvents[event.UserID] = append(r.userEvents[event.UserID], event.ID)

	return nil
}

func (r *EventRepository) Update(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.events[event.ID]
	if !exists {
		return domain.ErrEventNotFound
	}

	if existing.UserID != event.UserID {
		return domain.ErrInvalidUser
	}

	event.CreatedAt = existing.CreatedAt
	event.UpdatedAt = time.Now()
	r.events[event.ID] = event

	return nil
}

func (r *EventRepository) Delete(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, exists := r.events[id]
	if !exists {
		return domain.ErrEventNotFound
	}

	if event.UserID != userID {
		return domain.ErrInvalidUser
	}

	delete(r.events, id)

	// Remove from user index
	userEvents := r.userEvents[userID]
	for i, eid := range userEvents {
		if eid == id {
			r.userEvents[userID] = append(userEvents[:i], userEvents[i+1:]...)
			break
		}
	}

	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, exists := r.events[id]
	if !exists {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

func (r *EventRepository) GetEventsForDay(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.filterEvents(userID, func(e *domain.Event) bool {
		return isSameDay(e.Date, date)
	}), nil
}

func (r *EventRepository) GetEventsForWeek(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startOfWeek := getStartOfWeek(date)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	return r.filterEvents(userID, func(e *domain.Event) bool {
		return (e.Date.Equal(startOfWeek) || e.Date.After(startOfWeek)) && e.Date.Before(endOfWeek)
	}), nil
}

func (r *EventRepository) GetEventsForMonth(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return r.filterEvents(userID, func(e *domain.Event) bool {
		return (e.Date.Equal(startOfMonth) || e.Date.After(startOfMonth)) && e.Date.Before(endOfMonth)
	}), nil
}

func (r *EventRepository) ArchiveOldEvents(ctx context.Context, beforeDate time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	archivedCount := 0
	toArchive := make([]string, 0)

	for id, event := range r.events {
		if event.Date.Before(beforeDate) {
			toArchive = append(toArchive, id)
			r.archived = append(r.archived, &domain.ArchivedEvent{
				Event:      *event,
				ArchivedAt: time.Now(),
			})
			archivedCount++
		}
	}

	for _, id := range toArchive {
		event := r.events[id]
		delete(r.events, id)

		userEvents := r.userEvents[event.UserID]
		for i, eid := range userEvents {
			if eid == id {
				r.userEvents[event.UserID] = append(userEvents[:i], userEvents[i+1:]...)
				break
			}
		}
	}

	return archivedCount, nil
}

func (r *EventRepository) GetEventsWithReminders(ctx context.Context) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.Event, 0)
	for _, event := range r.events {
		if event.NeedsReminder() {
			events = append(events, event)
		}
	}
	return events, nil
}

func (r *EventRepository) Ping(ctx context.Context) error {
	return nil // Memory repository always available
}

// Helper methods
func (r *EventRepository) filterEvents(userID string, predicate func(*domain.Event) bool) []*domain.Event {
	events := make([]*domain.Event, 0)
	eventIDs := r.userEvents[userID]

	for _, id := range eventIDs {
		if event, exists := r.events[id]; exists && predicate(event) {
			events = append(events, event)
		}
	}

	return events
}

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func getStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1)
}
