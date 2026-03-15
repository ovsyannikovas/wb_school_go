package notify

import (
	"wb_school/L3/L3.1/backend/internal/infrastructure/rabbitmq"
	"wb_school/L3/L3.1/backend/internal/infrastructure/redis"
)

type NotifyHandlers struct {
	cache *redis.Client
	queue *rabbitmq.Queue
}

func SetupHandlers(cache *redis.Client, queue *rabbitmq.Queue) *NotifyHandlers {
	return &NotifyHandlers{
		cache: cache,
		queue: queue,
	}
}
