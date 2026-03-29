package domain

import (
	"strings"
	"time"
)

type Analytics struct {
	ID         int64     `json:"id" db:"id"`
	ShortCode  string    `json:"short_code" db:"short_code"`
	AccessedAt time.Time `json:"accessed_at" db:"accessed_at"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	IP         string    `json:"ip" db:"ip_address"`
	Referer    string    `json:"referer" db:"referer"`
	DeviceType string    `json:"device_type" db:"device_type"`
}

func ParseDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "mobile"):
		return "mobile"
	case strings.Contains(ua, "tablet"):
		return "tablet"
	case strings.Contains(ua, "windows") || strings.Contains(ua, "mac") || strings.Contains(ua, "linux"):
		return "desktop"
	default:
		return "other"
	}
}
