package notify

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"wb_school/L3/L3.1/backend/internal/domain"
)

func (h *NotifyHandlers) CancelNotification(w http.ResponseWriter, r *http.Request) {
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

	if notification.Status != domain.StatusPending {
		http.Error(w, "Cannot cancel non-pending notification", http.StatusBadRequest)
		return
	}

	notification.Status = domain.StatusCancelled
	notification.UpdatedAt = time.Now()

	if err := h.cache.SetNotification(notification); err != nil {
		http.Error(w, "Failed to cancel notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
