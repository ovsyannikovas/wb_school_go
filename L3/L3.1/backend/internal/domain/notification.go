package domain

import (
	"encoding/json"
	"time"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

type Notification struct {
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	Channel    string             `json:"channel"`
	Recipient  string             `json:"recipient"`
	Content    string             `json:"content"`
	SendAt     time.Time          `json:"send_at"`
	Status     NotificationStatus `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	RetryCount int                `json:"retry_count"`
	LastError  string             `json:"last_error,omitempty"`
}

func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Notification) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *Notification) IsDue() bool {
	return time.Now().After(n.SendAt) || time.Now().Equal(n.SendAt)
}
