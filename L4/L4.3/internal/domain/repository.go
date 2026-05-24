package domain

import (
	"context"
	"time"
)

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id, userID string) error
	GetByID(ctx context.Context, id string) (*Event, error)

	GetEventsForDay(ctx context.Context, userID string, date time.Time) ([]*Event, error)
	GetEventsForWeek(ctx context.Context, userID string, date time.Time) ([]*Event, error)
	GetEventsForMonth(ctx context.Context, userID string, date time.Time) ([]*Event, error)

	ArchiveOldEvents(ctx context.Context, beforeDate time.Time) (int, error)

	GetEventsWithReminders(ctx context.Context) ([]*Event, error)

	Ping(ctx context.Context) error
}

type ArchivedEventRepository interface {
	SaveArchived(ctx context.Context, event *ArchivedEvent) error
	GetArchived(ctx context.Context, userID string, limit, offset int) ([]*ArchivedEvent, error)
}
