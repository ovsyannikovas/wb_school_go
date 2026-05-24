package domain

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidDate   = errors.New("invalid date")
	ErrInvalidUser   = errors.New("invalid user id")
	ErrInvalidEvent  = errors.New("invalid event data")
	ErrEventConflict = errors.New("event conflict")
)

type Event struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Date        time.Time  `json:"date"`
	Event       string     `json:"event"`
	HasReminder bool       `json:"has_reminder"`
	ReminderAt  *time.Time `json:"reminder_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ArchivedEvent struct {
	Event
	ArchivedAt time.Time `json:"archived_at"`
}

func (e *Event) Validate() error {
	if e.UserID == "" {
		return ErrInvalidUser
	}
	if e.Event == "" {
		return errors.New("event text cannot be empty")
	}
	if e.Date.IsZero() {
		return ErrInvalidDate
	}
	if e.HasReminder && e.ReminderAt == nil {
		return errors.New("reminder time required when has_reminder is true")
	}
	if e.ReminderAt != nil && e.ReminderAt.Before(time.Now()) {
		return errors.New("reminder time cannot be in the past")
	}
	return nil
}

func (e *Event) NeedsReminder() bool {
	return e.HasReminder && e.ReminderAt != nil && e.ReminderAt.After(time.Now())
}
