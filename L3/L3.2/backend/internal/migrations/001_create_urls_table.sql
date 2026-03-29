-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
                                    id VARCHAR(36) PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
                             user_id VARCHAR(36),
    is_custom BOOLEAN DEFAULT FALSE,
    clicks INTEGER DEFAULT 0
    );

CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_urls_created_at ON urls(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd