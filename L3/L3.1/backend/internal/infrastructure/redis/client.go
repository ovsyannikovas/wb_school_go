package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"wb_school/L3/L3.1/backend/internal/domain"
)

type Client struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

func NewClient(addr, password string, db int) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
		ttl:    24 * time.Hour,
	}, nil
}

func (c *Client) SetNotification(notification *domain.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	key := fmt.Sprintf("notification:%s", notification.ID)
	if err := c.client.Set(c.ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	statusKey := fmt.Sprintf("status:%s:%s", notification.Status, notification.ID)
	c.client.Set(c.ctx, statusKey, notification.ID, c.ttl)

	if notification.Status == domain.StatusPending {
		scheduleKey := fmt.Sprintf("schedule:%d:%s", notification.SendAt.Unix(), notification.ID)
		c.client.Set(c.ctx, scheduleKey, notification.ID, c.ttl)
	}

	return nil
}

func (c *Client) GetNotification(id string) (*domain.Notification, error) {
	key := fmt.Sprintf("notification:%s", id)
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	var notification domain.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	return &notification, nil
}

func (c *Client) DeleteNotification(id string) error {
	key := fmt.Sprintf("notification:%s", id)

	notification, _ := c.GetNotification(id)
	if notification != nil {
		statusKey := fmt.Sprintf("status:%s:%s", notification.Status, id)
		c.client.Del(c.ctx, statusKey)

		if notification.Status == domain.StatusPending {
			scheduleKey := fmt.Sprintf("schedule:%d:%s", notification.SendAt.Unix(), id)
			c.client.Del(c.ctx, scheduleKey)
		}
	}

	return c.client.Del(c.ctx, key).Err()
}

func (c *Client) ListByStatus(status domain.NotificationStatus) ([]*domain.Notification, error) {
	pattern := fmt.Sprintf("status:%s:*", status)
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var notifications []*domain.Notification
	for _, key := range keys {
		id, err := c.client.Get(c.ctx, key).Result()
		if err != nil {
			continue
		}
		notification, err := c.GetNotification(id)
		if err != nil {
			continue
		}
		if notification != nil {
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
