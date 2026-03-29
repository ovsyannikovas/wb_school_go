package shortener

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"wb_school/L3/L3.2/backend/internal"
	"wb_school/L3/L3.2/backend/internal/domain"
	"wb_school/L3/L3.2/backend/internal/service"
)

type Handlers struct {
	urlRepo       service.URLRepository
	analyticsRepo service.AnalyticsRepository
	cache         service.CacheService
	generator     service.GeneratorService
	config        *internal.Config
	logger        *log.Logger
}

type CreateURLRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
	ExpiresIn  *int   `json:"expires_in,omitempty"`
}

type CreateURLResponse struct {
	ShortURL    string `json:"short_url"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHandlers(
	urlRepo service.URLRepository,
	analyticsRepo service.AnalyticsRepository,
	cache service.CacheService,
	generator service.GeneratorService,
	config *internal.Config,
	logger *log.Logger,
) *Handlers {
	if logger == nil {
		logger = log.Default()
	}
	return &Handlers{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		cache:         cache,
		generator:     generator,
		config:        config,
		logger:        logger,
	}
}

func (h *Handlers) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("ERROR: Failed to decode request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		h.logger.Printf("ERROR: URL is empty in request")
		respondWithError(w, http.StatusBadRequest, "URL is required")
		return
	}

	var shortCode string
	var isCustom bool

	if req.CustomCode != "" {
		if !h.generator.ValidateCustomCode(req.CustomCode) {
			h.logger.Printf("ERROR: Invalid custom code format: %s", req.CustomCode)
			respondWithError(w, http.StatusBadRequest, "Invalid custom code. Use 3-20 chars, letters and digits only")
			return
		}

		existing, err := h.urlRepo.GetByShortCode(r.Context(), req.CustomCode)
		if err != nil {
			h.logger.Printf("ERROR: Failed to check existing custom code: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if existing != nil {
			h.logger.Printf("ERROR: Custom code already taken: %s", req.CustomCode)
			respondWithError(w, http.StatusConflict, "Custom code already taken")
			return
		}

		shortCode = req.CustomCode
		isCustom = true
	} else {
		for i := 0; i < 10; i++ {
			code := h.generator.GenerateShortCode(h.config.ShortLength)
			existing, err := h.urlRepo.GetByShortCode(r.Context(), code)
			if err != nil {
				h.logger.Printf("WARNING: Failed to check code availability (attempt %d): %v", i+1, err)
				continue
			}
			if existing == nil {
				shortCode = code
				break
			}
		}
		if shortCode == "" {
			errMsg := "Failed to generate unique code after 10 attempts"
			h.logger.Printf("ERROR: %s", errMsg)
			respondWithError(w, http.StatusInternalServerError, "Failed to generate unique code")
			return
		}
	}

	url := &domain.ShortURL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		IsCustom:    isCustom,
		CreatedAt:   time.Now(),
		Clicks:      0,
	}

	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		url.ExpiresAt = &expiresAt
	}

	if err := h.urlRepo.Create(r.Context(), url); err != nil {
		h.logger.Printf("ERROR: Failed to create short URL in database: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	h.logger.Printf("INFO: Successfully created short URL: %s -> %s", shortCode, req.URL)
	response := CreateURLResponse{
		ShortURL:    h.config.BaseURL + "/s/" + shortCode,
		ShortCode:   shortCode,
		OriginalURL: req.URL,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["short_url"]

	url, err := h.cache.Get(r.Context(), shortCode)
	if err != nil {
		h.logger.Printf("WARNING: Cache get error for %s: %v", shortCode, err)
	}

	if url == nil {
		url, err = h.urlRepo.GetByShortCode(r.Context(), shortCode)
		if err != nil {
			h.logger.Printf("ERROR: Database error getting URL %s: %v", shortCode, err)
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if url == nil {
			h.logger.Printf("INFO: URL not found: %s", shortCode)
			respondWithError(w, http.StatusNotFound, "URL not found")
			return
		}

		if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
			h.logger.Printf("WARNING: Expired URL accessed: %s (expired at %s)", shortCode, url.ExpiresAt)
			respondWithError(w, http.StatusGone, "URL has expired")
			return
		}

		if err := h.cache.Set(r.Context(), shortCode, url); err != nil {
			h.logger.Printf("WARNING: Failed to cache URL %s: %v", shortCode, err)
		}
	}

	// СОЗДАЕМ НОВЫЙ КОНТЕКСТ ДЛЯ ФОНОВЫХ ЗАДАЧ
	bgCtx := context.Background()

	go func() {
		if err := h.saveAnalytics(bgCtx, r, shortCode); err != nil {
			h.logger.Printf("ERROR: Failed to save analytics for %s: %v", shortCode, err)
		}
	}()

	go func() {
		if err := h.urlRepo.IncrementClicks(bgCtx, shortCode); err != nil {
			h.logger.Printf("ERROR: Failed to increment clicks for %s: %v", shortCode, err)
		}
	}()

	h.logger.Printf("INFO: Redirected %s to %s", shortCode, url.OriginalURL)
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func (h *Handlers) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["short_url"]

	url, err := h.urlRepo.GetByShortCode(r.Context(), shortCode)
	if err != nil {
		h.logger.Printf("ERROR: Database error getting URL for analytics %s: %v", shortCode, err)
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if url == nil {
		h.logger.Printf("INFO: Analytics requested for non-existent URL: %s", shortCode)
		respondWithError(w, http.StatusNotFound, "URL not found")
		return
	}

	stats, err := h.analyticsRepo.GetStats(r.Context(), shortCode)
	if err != nil {
		h.logger.Printf("ERROR: Failed to get stats for %s: %v", shortCode, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}

	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()
	daily, err := h.analyticsRepo.GetStatsByPeriod(r.Context(), shortCode, from, to)
	if err != nil {
		h.logger.Printf("ERROR: Failed to get daily stats for %s: %v", shortCode, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get daily stats")
		return
	}

	response := map[string]interface{}{
		"stats": stats,
		"daily": daily,
		"url":   url,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) saveAnalytics(ctx context.Context, r *http.Request, shortCode string) error {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	if strings.HasPrefix(ip, "[") && strings.Contains(ip, "]") {
		if idx := strings.LastIndex(ip, "]"); idx != -1 {
			ip = ip[1:idx]
		}
	}

	analytics := &domain.Analytics{
		ShortCode:  shortCode,
		AccessedAt: time.Now(),
		UserAgent:  r.UserAgent(),
		IP:         ip,
		Referer:    r.Referer(),
		DeviceType: domain.ParseDeviceType(r.UserAgent()),
	}

	if err := h.analyticsRepo.Save(ctx, analytics); err != nil {
		return err
	}
	return nil
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR: Failed to encode JSON response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, ErrorResponse{Error: message})
}
