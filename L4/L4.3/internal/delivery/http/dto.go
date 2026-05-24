package http

import (
	"time"

	"L4_3/internal/domain"
)

type CreateEventRequest struct {
	UserID      string `json:"user_id"`
	Date        string `json:"date"`
	Event       string `json:"event"`
	HasReminder bool   `json:"has_reminder"`
	ReminderAt  string `json:"reminder_at,omitempty"`
}

type UpdateEventRequest struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Event       *string `json:"event,omitempty"`
	Date        *string `json:"date,omitempty"`
	HasReminder *bool   `json:"has_reminder,omitempty"`
	ReminderAt  *string `json:"reminder_at,omitempty"`
}

type DeleteEventRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type EventResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Date        string    `json:"date"`
	Event       string    `json:"event"`
	HasReminder bool      `json:"has_reminder"`
	ReminderAt  *string   `json:"reminder_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toEventResponse(e *domain.Event) EventResponse {
	resp := EventResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		Date:        e.Date.Format("2006-01-02"),
		Event:       e.Event,
		HasReminder: e.HasReminder,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	if e.ReminderAt != nil {
		reminderStr := e.ReminderAt.Format("2006-01-02 15:04:05")
		resp.ReminderAt = &reminderStr
	}

	return resp
}
