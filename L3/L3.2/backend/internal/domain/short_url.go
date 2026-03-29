package domain

import (
	"time"
)

type ShortURL struct {
	ID          string     `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	UserID      string     `json:"user_id,omitempty" db:"user_id"`
	IsCustom    bool       `json:"is_custom" db:"is_custom"`
	Clicks      int        `json:"clicks" db:"clicks"`
}
