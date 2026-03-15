package notify

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"wb_school/L3/L3.1/backend/internal/domain"
)

type CreateNotificationRequest struct {
	UserID    string    `json:"user_id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	SendAt    time.Time `json:"send_at"`
}

func (h *NotifyHandlers) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Recipient == "" || req.Content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if req.Channel != "email" && req.Channel != "telegram" && req.Channel != "sms" {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}

	if req.SendAt.Before(time.Now()) {
		http.Error(w, "Send time must be in the future", http.StatusBadRequest)
		return
	}

	now := time.Now()
	notification := &domain.Notification{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		Channel:    req.Channel,
		Recipient:  req.Recipient,
		Content:    req.Content,
		SendAt:     req.SendAt,
		Status:     domain.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		RetryCount: 0,
	}

	if err := h.cache.SetNotification(notification); err != nil {
		http.Error(w, "Failed to save notification", http.StatusInternalServerError)
		return
	}

	if err := h.queue.Publish(notification); err != nil {
		err := h.cache.DeleteNotification(notification.ID)
		if err != nil {
			http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		}
		http.Error(w, "Failed to schedule notification", http.StatusInternalServerError)
		return
	}

	response := NotificationResponse{
		ID:        notification.ID,
		Status:    string(notification.Status),
		SendAt:    notification.SendAt,
		CreatedAt: notification.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
