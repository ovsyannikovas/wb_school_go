package internal

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	RabbitMQURL    string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	MaxRetries     int
	RetryDelayBase time.Duration
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {

	}

	maxRetries, _ := strconv.Atoi(getEnv("MAX_RETRIES", "3"))
	retryDelayBase, _ := strconv.Atoi(getEnv("RETRY_DELAY_BASE", "5"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
		MaxRetries:     maxRetries,
		RetryDelayBase: time.Duration(retryDelayBase) * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
