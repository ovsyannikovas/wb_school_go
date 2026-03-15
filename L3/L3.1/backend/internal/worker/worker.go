package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/streadway/amqp"
	"wb_school/L3/L3.1/backend/internal/domain"
	"wb_school/L3/L3.1/backend/internal/infrastructure/rabbitmq"
	"wb_school/L3/L3.1/backend/internal/infrastructure/redis"
)

type Worker struct {
	cache          *redis.Client
	queue          *rabbitmq.Queue
	maxRetries     int
	retryDelayBase time.Duration
	stopChan       chan struct{}
}

func NewWorker(
	cache *redis.Client,
	queue *rabbitmq.Queue,
	maxRetries int,
	retryDelayBase time.Duration,
) *Worker {
	return &Worker{
		cache:          cache,
		queue:          queue,
		maxRetries:     maxRetries,
		retryDelayBase: retryDelayBase,
		stopChan:       make(chan struct{}),
	}
}

func (w *Worker) Start() error {
	msgs, err := w.queue.Consume()
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Consumer channel closed")
					return
				}
				w.processMessage(msg)
			}
		}
	}()

	log.Println("Worker started")
	return nil
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) processMessage(msg amqp.Delivery) {
	var queueMsg rabbitmq.QueueMessage
	if err := json.Unmarshal(msg.Body, &queueMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false)
		return
	}

	notification, err := w.cache.GetNotification(queueMsg.NotificationID)
	if err != nil {
		log.Printf("Failed to get notification %s: %v", queueMsg.NotificationID, err)
		msg.Nack(false, true)
		return
	}

	if notification == nil {
		log.Printf("Notification %s not found", queueMsg.NotificationID)
		msg.Ack(false)
		return
	}

	// Дополнительная проверка времени (на случай рассинхронизации)
	if time.Now().Before(notification.SendAt) {
		log.Printf("Warning: Notification %s received before send time. Rescheduling...", notification.ID)

		// Публикуем заново с оставшейся задержкой
		w.queue.Publish(notification)
		msg.Ack(false)
		return
	}

	if notification.Status != domain.StatusPending {
		log.Printf("Notification %s is not pending (status: %s)", notification.ID, notification.Status)
		msg.Ack(false)
		return
	}

	if err := w.sendNotification(notification); err != nil {
		log.Printf("Failed to send notification %s: %v", notification.ID, err)
		w.handleFailure(notification, err)

		if notification.RetryCount < w.maxRetries {
			w.scheduleRetry(notification)
		}

		msg.Ack(false)
		return
	}

	notification.Status = domain.StatusSent
	notification.UpdatedAt = time.Now()
	w.cache.SetNotification(notification)

	log.Printf("Notification %s sent successfully", notification.ID)
	msg.Ack(false)
}

func (w *Worker) sendNotification(notification *domain.Notification) error {
	switch notification.Channel {
	case "email":
		return w.sendEmail(notification)
	case "telegram":
		return w.sendTelegram(notification)
	case "sms":
		return w.sendSMS(notification)
	default:
		return fmt.Errorf("unknown channel: %s", notification.Channel)
	}
}

func (w *Worker) sendEmail(notification *domain.Notification) error {
	log.Printf("Sending email to %s: %s", notification.Recipient, notification.Content)
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (w *Worker) sendTelegram(notification *domain.Notification) error {
	log.Printf("Sending Telegram message to %s: %s", notification.Recipient, notification.Content)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (w *Worker) sendSMS(notification *domain.Notification) error {
	log.Printf("Sending SMS to %s: %s", notification.Recipient, notification.Content)
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (w *Worker) handleFailure(notification *domain.Notification, err error) {
	notification.RetryCount++
	notification.LastError = err.Error()
	notification.UpdatedAt = time.Now()

	if notification.RetryCount >= w.maxRetries {
		notification.Status = domain.StatusFailed
	}

	w.cache.SetNotification(notification)
}

func (w *Worker) scheduleRetry(notification *domain.Notification) {
	// base * 2^(retry-1)
	delay := w.retryDelayBase * time.Duration(math.Pow(2, float64(notification.RetryCount-1)))

	notification.SendAt = time.Now().Add(delay)
	notification.UpdatedAt = time.Now()

	if err := w.queue.Publish(notification); err != nil {
		log.Printf("Failed to schedule retry for %s: %v", notification.ID, err)
	}

	log.Printf("Scheduled retry for %s in %v (attempt %d/%d)",
		notification.ID, delay, notification.RetryCount, w.maxRetries)
}
