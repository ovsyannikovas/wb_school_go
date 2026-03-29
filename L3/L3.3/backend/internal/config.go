package internal

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	PostgresURL string
	BaseURL     string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/comment?sslmode=disable"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
