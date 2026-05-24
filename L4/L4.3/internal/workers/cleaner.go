package workers

import (
	"context"
	"L4_3/internal/usecase"
	"L4_3/pkg/logger"
	"time"
)

type CleanerWorker struct {
	eventUsecase *usecase.EventUsecase
	logger       *logger.AsyncLogger
	interval     time.Duration
	stop         chan struct{}
}

func NewCleanerWorker(usecase *usecase.EventUsecase, log *logger.AsyncLogger, interval time.Duration) *CleanerWorker {
	return &CleanerWorker{
		eventUsecase: usecase,
		logger:       log,
		interval:     interval,
		stop:         make(chan struct{}),
	}
}

func (w *CleanerWorker) Start() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				archivedCount, err := w.eventUsecase.ArchiveOldEvents(ctx, 30)
				if err != nil {
					w.logger.Error("Failed to archive old events", map[string]interface{}{
						"error": err.Error(),
					})
				} else if archivedCount > 0 {
					w.logger.Info("Archived old events", map[string]interface{}{
						"count": archivedCount,
					})
				}
			case <-w.stop:
				return
			}
		}
	}()
}

func (w *CleanerWorker) Stop() {
	close(w.stop)
}
