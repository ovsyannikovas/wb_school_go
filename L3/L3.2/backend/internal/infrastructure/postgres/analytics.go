package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"wb_school/L3/L3.2/backend/internal/domain"
	"wb_school/L3/L3.2/backend/internal/service"
)

type AnalyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) Save(ctx context.Context, analytics *domain.Analytics) error {
	query := `
		INSERT INTO analytics (short_code, accessed_at, user_agent, ip_address, referer, device_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		analytics.ShortCode, analytics.AccessedAt, analytics.UserAgent,
		analytics.IP, analytics.Referer, analytics.DeviceType,
	)

	return err
}

func (r *AnalyticsRepository) GetByShortCode(ctx context.Context, shortCode string) ([]*domain.Analytics, error) {
	var analytics []*domain.Analytics
	query := `SELECT * FROM analytics WHERE short_code = $1 ORDER BY accessed_at DESC`
	err := r.db.SelectContext(ctx, &analytics, query, shortCode)
	return analytics, err
}

func (r *AnalyticsRepository) GetStats(ctx context.Context, shortCode string) (*service.URLAnalytics, error) {
	stats := &service.URLAnalytics{ShortCode: shortCode}

	err := r.db.GetContext(ctx, &stats.TotalClicks,
		`SELECT COUNT(*) FROM analytics WHERE short_code = $1`, shortCode)
	if err != nil {
		return nil, err
	}

	err = r.db.GetContext(ctx, &stats.UniqueIPs,
		`SELECT COUNT(DISTINCT ip_address) FROM analytics WHERE short_code = $1 AND ip_address IS NOT NULL`, shortCode)
	if err != nil {
		return nil, err
	}

	type DeviceCount struct {
		DeviceType string `db:"device_type"`
		Count      int    `db:"count"`
	}

	var deviceCounts []DeviceCount
	err = r.db.SelectContext(ctx, &deviceCounts,
		`SELECT device_type, COUNT(*) as count FROM analytics WHERE short_code = $1 GROUP BY device_type`, shortCode)
	if err != nil {
		return nil, err
	}

	for _, dc := range deviceCounts {
		switch dc.DeviceType {
		case "desktop":
			stats.DesktopClicks = dc.Count
		case "mobile":
			stats.MobileClicks = dc.Count
		default:
			stats.OtherClicks += dc.Count
		}
	}

	return stats, nil
}

func (r *AnalyticsRepository) GetStatsByPeriod(ctx context.Context, shortCode string, from, to time.Time) ([]*service.DailyStats, error) {
	var stats []*service.DailyStats
	query := `
		SELECT DATE(accessed_at) as date, COUNT(*) as clicks
		FROM analytics
		WHERE short_code = $1 AND accessed_at BETWEEN $2 AND $3
		GROUP BY DATE(accessed_at)
		ORDER BY date DESC
	`

	err := r.db.SelectContext(ctx, &stats, query, shortCode, from, to)
	return stats, err
}
