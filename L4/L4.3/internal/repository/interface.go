package repository

import (
	"L4_3/internal/domain"
)

// DB represents a database for health check
type DB interface {
	Ping() error
}

// Repos holds all repositories
type Repos struct {
	Events   domain.EventRepository
	Archived domain.ArchivedEventRepository
	DB       DB
}
