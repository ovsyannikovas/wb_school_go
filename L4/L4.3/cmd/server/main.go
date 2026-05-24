package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"L4_3/config"
	delivery "L4_3/internal/delivery/http"
	"L4_3/internal/middleware"
	"L4_3/internal/repository/memory"
	"L4_3/internal/usecase"
	"L4_3/internal/workers"
	"L4_3/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log, err := logger.New(logger.Config{
		BufferSize: cfg.LoggerBufferSize,
		Output:     os.Stdout,
		TimeFormat: "2006-01-02 15:04:05.000",
	})
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer log.Close()

	log.Info("Starting Calendar Service", map[string]interface{}{
		"port": cfg.Port,
	})

	// Initialize repository
	eventRepo := memory.NewEventRepository()

	// Initialize usecase
	eventUsecase := usecase.NewEventUsecase(eventRepo)

	// Initialize workers
	reminderWorker := workers.NewReminderWorker(eventUsecase, log, cfg.ReminderCheckInterval)
	cleanerWorker := workers.NewCleanerWorker(eventUsecase, log, cfg.CleanerInterval)

	// Start workers
	reminderWorker.Start()
	cleanerWorker.Start()
	defer reminderWorker.Stop()
	defer cleanerWorker.Stop()

	// Initialize HTTP handlers
	eventHandler := delivery.NewEventHandler(eventUsecase, log)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/create_event", eventHandler.CreateEvent)
	mux.HandleFunc("/update_event", eventHandler.UpdateEvent)
	mux.HandleFunc("/delete_event", eventHandler.DeleteEvent)
	mux.HandleFunc("/events_for_day", eventHandler.GetEventsForDay)
	mux.HandleFunc("/events_for_week", eventHandler.GetEventsForWeek)
	mux.HandleFunc("/events_for_month", eventHandler.GetEventsForMonth)

	// Apply middleware
	handler := middleware.LoggingMiddleware(log)(mux)

	// Create server
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server starting", map[string]interface{}{
			"address": cfg.Port,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...", nil)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	log.Info("Server exited", nil)
}
