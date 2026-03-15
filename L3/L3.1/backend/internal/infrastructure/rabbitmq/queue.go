package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"wb_school/L3/L3.1/backend/internal/domain"
)

type Queue struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	delayedQueue amqp.Queue // Очередь с TTL (сюда публикуем)
	targetQueue  amqp.Queue // Целевая очередь (отсюда читаем)
}

type QueueMessage struct {
	NotificationID string    `json:"notification_id"`
	SendAt         time.Time `json:"send_at"`
	RetryCount     int       `json:"retry_count"`
}

func NewQueue(url string) (*Queue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		"delayed_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	targetQueue, err := ch.QueueDeclare(
		"target_notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare target queue: %w", err)
	}

	err = ch.QueueBind(
		targetQueue.Name,
		"delayed_key",
		"delayed_exchange",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind target queue: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "delayed_exchange",
		"x-dead-letter-routing-key": "delayed_key",
	}

	delayedQueue, err := ch.QueueDeclare(
		"delayed_queue",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare delayed queue: %w", err)
	}

	log.Printf("RabbitMQ initialized: delayed_queue -> (TTL) -> delayed_exchange -> target_queue")

	return &Queue{
		conn:         conn,
		channel:      ch,
		delayedQueue: delayedQueue,
		targetQueue:  targetQueue,
	}, nil
}

func (q *Queue) Publish(notification *domain.Notification) error {
	msg := QueueMessage{
		NotificationID: notification.ID,
		SendAt:         notification.SendAt,
		RetryCount:     notification.RetryCount,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	delay := time.Until(notification.SendAt)
	if delay < 0 {
		delay = 0
	}

	err = q.channel.Publish(
		"",
		q.delayedQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Expiration:   fmt.Sprintf("%d", delay.Milliseconds()),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Notification %s scheduled for %v (delay: %v)",
		notification.ID, notification.SendAt, delay)
	return nil
}

func (q *Queue) Consume() (<-chan amqp.Delivery, error) {
	return q.channel.Consume(
		q.targetQueue.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

func (q *Queue) Close() error {
	if err := q.channel.Close(); err != nil {
		return err
	}
	return q.conn.Close()
}
