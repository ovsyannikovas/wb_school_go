package usecase

import (
	"context"
	"time"

	"L4_3/internal/domain"
)

type EventUsecase struct {
	eventRepo domain.EventRepository
}

func NewEventUsecase(repo domain.EventRepository) *EventUsecase {
	return &EventUsecase{
		eventRepo: repo,
	}
}

// CreateEvent creates a new event
func (u *EventUsecase) CreateEvent(ctx context.Context, userID, eventText string, date time.Time, hasReminder bool, reminderAt *time.Time) (*domain.Event, error) {
	event := &domain.Event{
		UserID:      userID,
		Event:       eventText,
		Date:        date,
		HasReminder: hasReminder,
		ReminderAt:  reminderAt,
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	if err := u.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// UpdateEvent updates an existing event
func (u *EventUsecase) UpdateEvent(ctx context.Context, id, userID string, eventText *string, date *time.Time, hasReminder *bool, reminderAt *time.Time) (*domain.Event, error) {
	existing, err := u.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.UserID != userID {
		return nil, domain.ErrInvalidUser
	}

	if eventText != nil {
		existing.Event = *eventText
	}
	if date != nil {
		existing.Date = *date
	}
	if hasReminder != nil {
		existing.HasReminder = *hasReminder
	}
	if reminderAt != nil {
		existing.ReminderAt = reminderAt
	}

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := u.eventRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteEvent deletes an event
func (u *EventUsecase) DeleteEvent(ctx context.Context, id, userID string) error {
	return u.eventRepo.Delete(ctx, id, userID)
}

// GetEventsForDay returns events for a specific day
func (u *EventUsecase) GetEventsForDay(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	return u.eventRepo.GetEventsForDay(ctx, userID, date)
}

// GetEventsForWeek returns events for a week
func (u *EventUsecase) GetEventsForWeek(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	return u.eventRepo.GetEventsForWeek(ctx, userID, date)
}

// GetEventsForMonth returns events for a month
func (u *EventUsecase) GetEventsForMonth(ctx context.Context, userID string, date time.Time) ([]*domain.Event, error) {
	return u.eventRepo.GetEventsForMonth(ctx, userID, date)
}

// GetEventByID returns an event by ID
func (u *EventUsecase) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	return u.eventRepo.GetByID(ctx, id)
}

// ArchiveOldEvents archives events older than specified date
func (u *EventUsecase) ArchiveOldEvents(ctx context.Context, daysOld int) (int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	return u.eventRepo.ArchiveOldEvents(ctx, cutoffDate)
}

// GetEventsWithReminders returns all events that need reminders
func (u *EventUsecase) GetEventsWithReminders(ctx context.Context) ([]*domain.Event, error) {
	return u.eventRepo.GetEventsWithReminders(ctx)
}
