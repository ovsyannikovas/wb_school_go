package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server config
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	// Database config
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Logger config
	LoggerBufferSize int

	// Worker config
	ReminderCheckInterval time.Duration
	CleanerInterval       time.Duration
	ArchivedDaysThreshold int
}

// Load reads configuration from environment variables
func Load() Config {
	return Config{
		Port:         getEnv("PORT", ":8080"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "calendar"),

		LoggerBufferSize: getEnvInt("LOGGER_BUFFER_SIZE", 1000),

		ReminderCheckInterval: getDurationEnv("REMINDER_CHECK_INTERVAL", 1*time.Minute),
		CleanerInterval:       getDurationEnv("CLEANER_INTERVAL", 1*time.Hour),
		ArchivedDaysThreshold: getEnvInt("ARCHIVED_DAYS_THRESHOLD", 30),
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt retrieves integer environment variable or returns default value
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// getDurationEnv retrieves duration environment variable or returns default value
func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	result, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return result
}
