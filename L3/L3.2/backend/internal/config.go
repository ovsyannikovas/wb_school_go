package internal

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	PostgresURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	BaseURL       string
	ShortLength   int
}

func Load() *Config {
	godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	shortLength, _ := strconv.Atoi(getEnv("SHORT_LENGTH", "6"))

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/comment?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		ShortLength:   shortLength,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
