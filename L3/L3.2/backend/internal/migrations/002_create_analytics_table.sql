-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analytics (
                                         id BIGSERIAL PRIMARY KEY,
                                         short_code VARCHAR(20) NOT NULL,
    accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET,
    referer TEXT,
    device_type VARCHAR(20) DEFAULT 'unknown',
    CONSTRAINT fk_analytics_short_code FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE
    );

CREATE INDEX idx_analytics_short_code ON analytics(short_code);
CREATE INDEX idx_analytics_accessed_at ON analytics(accessed_at);
CREATE INDEX idx_analytics_device_type ON analytics(device_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analytics;
-- +goose StatementEnd