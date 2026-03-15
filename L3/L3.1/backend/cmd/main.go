package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"wb_school/L3/L3.1/backend/internal"
	notifyhandler "wb_school/L3/L3.1/backend/internal/application/notify"
	"wb_school/L3/L3.1/backend/internal/infrastructure/rabbitmq"
	"wb_school/L3/L3.1/backend/internal/infrastructure/redis"
	"wb_school/L3/L3.1/backend/internal/worker"
)

func main() {
	cfg := internal.Load()

	redisClient, err := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	rabbitClient, err := rabbitmq.NewQueue(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	notifyWorker := worker.NewWorker(
		redisClient,
		rabbitClient,
		cfg.MaxRetries,
		cfg.RetryDelayBase,
	)
	if err := notifyWorker.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
	defer notifyWorker.Stop()

	handlers := notifyhandler.SetupHandlers(redisClient, rabbitClient)

	router := mux.NewRouter()

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	router.Use(corsMiddleware)

	router.HandleFunc("/api/notify", handlers.CreateNotification).Methods("POST")
	router.HandleFunc("/api/notify/{id}", handlers.GetNotification).Methods("GET")
	router.HandleFunc("/api/notify/{id}", handlers.CancelNotification).Methods("DELETE")
	router.HandleFunc("/api/notifications", handlers.ListNotifications).Methods("GET")

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
