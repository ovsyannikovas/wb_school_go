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
	"wb_school/L3/L3.2/backend/internal"
	"wb_school/L3/L3.2/backend/internal/application/shortener"
	"wb_school/L3/L3.2/backend/internal/infrastructure/postgres"
	"wb_school/L3/L3.2/backend/internal/infrastructure/redis"
	"wb_school/L3/L3.2/backend/lib/generator"
)

func main() {
	cfg := internal.Load()

	db, err := postgres.NewConnection(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	cache, err := redis.NewCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	urlRepo := postgres.NewURLRepository(db.DB)
	analyticsRepo := postgres.NewAnalyticsRepository(db.DB)
	urlGenerator := generator.NewGenerator()

	logger := log.New(os.Stdout, "[API] ", log.LstdFlags|log.Lshortfile)

	handlers := shortener.NewHandlers(urlRepo, analyticsRepo, cache, urlGenerator, cfg, logger)

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.Use(loggingMiddleware)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/shorten", handlers.CreateShortURL).Methods("POST", "OPTIONS")
	api.HandleFunc("/analytics/{short_url}", handlers.GetAnalytics).Methods("GET", "OPTIONS")

	r.HandleFunc("/s/{short_url}", handlers.Redirect).Methods("GET")

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
