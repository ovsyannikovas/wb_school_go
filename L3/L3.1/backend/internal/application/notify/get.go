package notify

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type NotificationResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	SendAt    time.Time `json:"send_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *NotifyHandlers) GetNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	notification, err := h.cache.GetNotification(id)
	if err != nil {
		http.Error(w, "Failed to get notification", http.StatusInternalServerError)
		return
	}

	if notification == nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}
