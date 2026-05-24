package http

import (
	"L4_3/internal/domain"
	"L4_3/internal/usecase"
	"L4_3/pkg/logger"
	"encoding/json"
	"net/http"
	"time"
)

type EventHandler struct {
	eventUsecase *usecase.EventUsecase
	logger       *logger.AsyncLogger
}

func NewEventHandler(usecase *usecase.EventUsecase, log *logger.AsyncLogger) *EventHandler {
	return &EventHandler{
		eventUsecase: usecase,
		logger:       log,
	}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	var reminderAt *time.Time
	if req.HasReminder && req.ReminderAt != "" {
		reminder, err := time.Parse("2006-01-02 15:04:05", req.ReminderAt)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid reminder_at format, use YYYY-MM-DD HH:MM:SS")
			return
		}
		reminderAt = &reminder
	}

	event, err := h.eventUsecase.CreateEvent(r.Context(), req.UserID, req.Event, date, req.HasReminder, reminderAt)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("Event created", map[string]interface{}{
		"event_id": event.ID,
		"user_id":  event.UserID,
	})

	h.sendSuccess(w, map[string]interface{}{
		"message": "Event created successfully",
		"event":   toEventResponse(event),
	})
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var date *time.Time
	if req.Date != nil {
		d, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
			return
		}
		date = &d
	}

	var reminderAt *time.Time
	if req.HasReminder != nil && *req.HasReminder && req.ReminderAt != nil {
		reminder, err := time.Parse("2006-01-02 15:04:05", *req.ReminderAt)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid reminder_at format, use YYYY-MM-DD HH:MM:SS")
			return
		}
		reminderAt = &reminder
	}

	event, err := h.eventUsecase.UpdateEvent(r.Context(), req.ID, req.UserID, req.Event, date, req.HasReminder, reminderAt)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("Event updated", map[string]interface{}{
		"event_id": event.ID,
		"user_id":  event.UserID,
	})

	h.sendSuccess(w, map[string]interface{}{
		"message": "Event updated successfully",
		"event":   toEventResponse(event),
	})
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var req DeleteEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.eventUsecase.DeleteEvent(r.Context(), req.ID, req.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("Event deleted", map[string]interface{}{
		"event_id": req.ID,
		"user_id":  req.UserID,
	})

	h.sendSuccess(w, map[string]interface{}{
		"message": "Event deleted successfully",
	})
}

func (h *EventHandler) GetEventsForDay(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, http.StatusBadRequest, "user_id and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	events, err := h.eventUsecase.GetEventsForDay(r.Context(), userID, date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := make([]EventResponse, len(events))
	for i, event := range events {
		response[i] = toEventResponse(event)
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": response,
		"count":  len(events),
	})
}

func (h *EventHandler) GetEventsForWeek(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, http.StatusBadRequest, "user_id and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	events, err := h.eventUsecase.GetEventsForWeek(r.Context(), userID, date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := make([]EventResponse, len(events))
	for i, event := range events {
		response[i] = toEventResponse(event)
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": response,
		"count":  len(events),
	})
}

func (h *EventHandler) GetEventsForMonth(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, http.StatusBadRequest, "user_id and date are required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	events, err := h.eventUsecase.GetEventsForMonth(r.Context(), userID, date)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := make([]EventResponse, len(events))
	for i, event := range events {
		response[i] = toEventResponse(event)
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": response,
		"count":  len(events),
	})
}

func (h *EventHandler) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": data,
	})
}

func (h *EventHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (h *EventHandler) handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrEventNotFound:
		h.sendError(w, http.StatusServiceUnavailable, err.Error())
	case domain.ErrInvalidDate, domain.ErrInvalidUser, domain.ErrInvalidEvent:
		h.sendError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		h.sendError(w, http.StatusInternalServerError, "Internal server error")
	}
}
