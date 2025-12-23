package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"latihan_cloud8/middleware"
	"latihan_cloud8/store"
	"latihan_cloud8/utils"
)

type NotificationHandler struct {
	Store *store.MySQLStore
}

func NewNotificationHandler(store *store.MySQLStore) *NotificationHandler {
	return &NotificationHandler{Store: store}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(middleware.UserCtxKey)
	claims := v.(*utils.Claims) // Safely assume AuthMiddleware ran

	user, err := h.Store.GetByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	notifs, err := h.Store.GetNotifications(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.MarkNotificationRead(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := h.Store.MarkNotificationRead(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteNotification(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SendNotification (Admin Only) - Broadcast or Direct
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID  string `json:"user_id"` // "all" for broadcast
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Message == "" {
		http.Error(w, "Message required", http.StatusBadRequest)
		return
	}

	if payload.UserID == "all" {
		// Broadcast logic
		users, _ := h.Store.GetAllUsers()
		for _, u := range users {
			h.Store.CreateNotification(u.ID, payload.Message)
		}
	} else {
		h.Store.CreateNotification(payload.UserID, payload.Message)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification sent"})
}

func (h *NotificationHandler) ShowNotificationsPage(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	claims := v.(*utils.Claims)

	utils.RenderTemplate(w, "notifications.html", map[string]interface{}{
		"Title":      "Notifikasi",
		"ActivePage": "notifications",
		"Username":   claims.Username,
		"Role":       claims.Role,
	})
}
