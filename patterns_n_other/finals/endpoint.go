package main

import (
	"context"
	"database/sql"
	"net/http"
	"sync/atomic"
	"time"
)

type Health struct {
	db       *sql.DB
	started  atomic.Bool
	stopping atomic.Bool
}

func NewHealth(db *sql.DB) *Health {
	return &Health{db: db}
}

// SetStarted вызывается после завершения инициализации (например, после миграций)
func (h *Health) SetStarted() {
	h.started.Store(true)
}

// SetStopping вызывается при graceful shutdown (получен сигнал SIGTERM)
func (h *Health) SetStopping() {
	h.stopping.Store(true)
}

func (h *Health) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

func (h *Health) Startup(w http.ResponseWriter, r *http.Request) {
	if !h.started.Load() {
		http.Error(w, `{"status":"starting"}`, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"started"}`))
}

func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	if !h.started.Load() {
		http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
		return
	}
	if h.stopping.Load() {
		http.Error(w, `{"status":"shutting down"}`, http.StatusServiceUnavailable)
		return
	}

	// Проверяем БД с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		http.Error(w, `{"status":"db unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func main() {
	db := &sql.DB{}
	health := NewHealth(db)
	health.SetStarted()

	mux := http.NewServeMux()
	mux.HandleFunc("/live", health.Live)
	mux.HandleFunc("/ready", health.Ready)
	mux.HandleFunc("/startup", health.Startup)

	// При graceful shutdown:
	health.SetStopping()
}
