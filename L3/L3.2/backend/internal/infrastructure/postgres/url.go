package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"wb_school/L3/L3.2/backend/internal/domain"
)

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, url *domain.ShortURL) error {
	if url.ID == "" {
		url.ID = uuid.New().String()
	}
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO urls (id, short_code, original_url, created_at, expires_at, user_id, is_custom, clicks)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		url.ID, url.ShortCode, url.OriginalURL, url.CreatedAt,
		url.ExpiresAt, url.UserID, url.IsCustom, url.Clicks,
	)

	return err
}

func (r *URLRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.ShortURL, error) {
	var url domain.ShortURL
	query := `SELECT * FROM urls WHERE short_code = $1`

	err := r.db.GetContext(ctx, &url, query, shortCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*domain.ShortURL, error) {
	var url domain.ShortURL
	query := `SELECT * FROM urls WHERE original_url = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	err := r.db.GetContext(ctx, &url, query, originalURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &url, err
}

func (r *URLRepository) Update(ctx context.Context, url *domain.ShortURL) error {
	query := `
		UPDATE urls 
		SET original_url = $1, expires_at = $2, user_id = $3, is_custom = $4, clicks = $5
		WHERE short_code = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		url.OriginalURL, url.ExpiresAt, url.UserID, url.IsCustom, url.Clicks, url.ShortCode,
	)

	return err
}

func (r *URLRepository) Delete(ctx context.Context, shortCode string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM urls WHERE short_code = $1`, shortCode)
	return err
}

func (r *URLRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`, shortCode)
	return err
}

func (r *URLRepository) List(ctx context.Context, limit, offset int) ([]*domain.ShortURL, error) {
	var urls []*domain.ShortURL
	query := `SELECT * FROM urls ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &urls, query, limit, offset)
	return urls, err
}
