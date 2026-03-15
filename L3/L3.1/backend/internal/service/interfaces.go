package service

import (
	"wb_school/L3/L3.1/backend/internal/domain"
)

type NotificationRepository interface {
	Save(notification *domain.Notification) error
	Get(id string) (*domain.Notification, error)
	Update(notification *domain.Notification) error
	Delete(id string) error
	ListByStatus(status domain.NotificationStatus) ([]*domain.Notification, error)
	GetDueNotifications(limit int) ([]*domain.Notification, error)
	Close() error
}

type QueueService interface {
	Publish(notification *domain.Notification) error
	Consume() (<-chan interface{}, error)
	Close() error
}

type NotifierService interface {
	Send(notification *domain.Notification) error
}

type CacheService interface {
	GetNotification(id string) (*domain.Notification, error)
	SetNotification(notification *domain.Notification) error
	DeleteNotification(id string) error
	ListByStatus(status domain.NotificationStatus) ([]*domain.Notification, error)
}
