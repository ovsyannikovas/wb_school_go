package middleware

import (
	"L4_3/pkg/logger"
	"encoding/json"
	"net/http"
	"time"
)

// LoggingMiddleware logs incoming HTTP requests and outgoing responses
func LoggingMiddleware(log *logger.AsyncLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrappedWriter, r)

			fields := map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"status":      wrappedWriter.statusCode,
				"duration":    time.Since(start).String(),
			}

			if wrappedWriter.statusCode >= 500 {
				log.Error("Request completed with error", fields)
			} else if wrappedWriter.statusCode >= 400 {
				log.Warn("Request completed with client error", fields)
			} else {
				log.Info("Request completed", fields)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LogEntry represents a single log entry for request/response
type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	Duration   string    `json:"duration"`
	RemoteAddr string    `json:"remote_addr"`
	UserAgent  string    `json:"user_agent"`
}

// WriteLog serializes log entry to JSON
func WriteLog(log *logger.AsyncLogger, entry LogEntry) {
	log.Info("HTTP Request", map[string]interface{}{
		"method":      entry.Method,
		"path":        entry.Path,
		"status":      entry.Status,
		"duration":    entry.Duration,
		"remote_addr": entry.RemoteAddr,
		"user_agent":  entry.UserAgent,
	})
}

// ValidateContentType checks if request has valid content type
func ValidateContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.ContentLength > 0 {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && contentType != "application/json" {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ResponseEncoder encodes response as JSON
type ResponseEncoder struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// EncodeJSON writes JSON response
func EncodeJSON(w http.ResponseWriter, status int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ResponseEncoder{
		Status:  status,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}
