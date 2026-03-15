package notify

import (
	"encoding/json"
	"net/http"

	"wb_school/L3/L3.1/backend/internal/domain"
)

func (h *NotifyHandlers) ListNotifications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = string(domain.StatusPending)
	}

	notifications, err := h.cache.ListByStatus(domain.NotificationStatus(status))
	if err != nil {
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
