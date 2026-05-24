package workers

import (
	"context"
	"L4_3/internal/domain"
	"L4_3/internal/usecase"
	"L4_3/pkg/logger"
	"time"
)

type ReminderWorker struct {
	eventUsecase *usecase.EventUsecase
	logger       *logger.AsyncLogger
	reminderCh   chan *domain.Event
	ticker       *time.Ticker
	stop         chan struct{}
}

func NewReminderWorker(usecase *usecase.EventUsecase, log *logger.AsyncLogger, checkInterval time.Duration) *ReminderWorker {
	return &ReminderWorker{
		eventUsecase: usecase,
		logger:       log,
		reminderCh:   make(chan *domain.Event, 100),
		ticker:       time.NewTicker(checkInterval),
		stop:         make(chan struct{}),
	}
}

func (w *ReminderWorker) Start() {
	go w.processReminders()
	go w.checkReminders()
}

func (w *ReminderWorker) checkReminders() {
	for {
		select {
		case <-w.ticker.C:
			ctx := context.Background()
			events, err := w.eventUsecase.GetEventsWithReminders(ctx)
			if err != nil {
				w.logger.Error("Failed to get events with reminders", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			now := time.Now()
			for _, event := range events {
				if event.ReminderAt != nil && event.ReminderAt.Before(now) {
					select {
					case w.reminderCh <- event:
					default:
						w.logger.Error("Reminder channel full, dropping reminder", map[string]interface{}{
							"event_id": event.ID,
						})
					}
				}
			}
		case <-w.stop:
			w.ticker.Stop()
			return
		}
	}
}

func (w *ReminderWorker) processReminders() {
	for event := range w.reminderCh {
		w.logger.Info("REMINDER", map[string]interface{}{
			"user_id":  event.UserID,
			"event":    event.Event,
			"date":     event.Date.Format("2006-01-02"),
			"event_id": event.ID,
		})
	}
}

func (w *ReminderWorker) Stop() {
	close(w.stop)
	w.ticker.Stop()
}
