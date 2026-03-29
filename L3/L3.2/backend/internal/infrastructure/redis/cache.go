package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"wb_school/L3/L3.2/backend/internal/domain"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
	ctx    context.Context
}

func NewCache(addr, password string, db int) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Cache{
		client: client,
		ttl:    1 * time.Hour,
		ctx:    ctx,
	}, nil
}

func (c *Cache) Get(ctx context.Context, shortCode string) (*domain.ShortURL, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var url domain.ShortURL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (c *Cache) Set(ctx context.Context, shortCode string, url *domain.ShortURL) error {
	key := fmt.Sprintf("url:%s", shortCode)
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
