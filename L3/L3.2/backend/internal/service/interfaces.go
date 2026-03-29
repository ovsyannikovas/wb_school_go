package service

import (
	"context"
	"time"
	"wb_school/L3/L3.2/backend/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.ShortURL) error
	GetByShortCode(ctx context.Context, shortCode string) (*domain.ShortURL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*domain.ShortURL, error)
	Update(ctx context.Context, url *domain.ShortURL) error
	Delete(ctx context.Context, shortCode string) error
	IncrementClicks(ctx context.Context, shortCode string) error
	List(ctx context.Context, limit, offset int) ([]*domain.ShortURL, error)
}

type AnalyticsRepository interface {
	Save(ctx context.Context, analytics *domain.Analytics) error
	GetByShortCode(ctx context.Context, shortCode string) ([]*domain.Analytics, error)
	GetStats(ctx context.Context, shortCode string) (*URLAnalytics, error)
	GetStatsByPeriod(ctx context.Context, shortCode string, from, to time.Time) ([]*DailyStats, error)
}

type CacheService interface {
	Get(ctx context.Context, shortCode string) (*domain.ShortURL, error)
	Set(ctx context.Context, shortCode string, url *domain.ShortURL) error
	Delete(ctx context.Context, shortCode string) error
}

type GeneratorService interface {
	GenerateShortCode(length int) string
	ValidateCustomCode(code string) bool
}

type URLAnalytics struct {
	ShortCode     string `json:"short_code"`
	TotalClicks   int    `json:"total_clicks"`
	UniqueIPs     int    `json:"unique_ips"`
	DesktopClicks int    `json:"desktop_clicks"`
	MobileClicks  int    `json:"mobile_clicks"`
	OtherClicks   int    `json:"other_clicks"`
}

type DailyStats struct {
	Date   string `json:"date" db:"date"`
	Clicks int    `json:"clicks" db:"clicks"`
}

type TopURL struct {
	ShortCode   string `json:"short_code" db:"short_code"`
	OriginalURL string `json:"original_url" db:"original_url"`
	Clicks      int    `json:"clicks" db:"clicks"`
}
