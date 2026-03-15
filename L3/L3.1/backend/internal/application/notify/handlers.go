package notify

import (
	"github.com/ovsyannikvas/wb_school/L3/L3.1/internal/infrastructure/redis"
	"github.com/ovsyannikvas/wb_school/L3/L3.1/internal/service"
)

type OrderHandlers struct {
	ordersRepo  service.OrderRepository
	redisClient *redis.Client
}

func SetupHandlers(
	or service.OrderRepository,
	rc *redis.Client,
) OrderHandlers {
	return OrderHandlers{
		ordersRepo:  or,
		redisClient: rc,
	}
}
