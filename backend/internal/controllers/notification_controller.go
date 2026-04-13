package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/helpers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// NotificationController handles HTTP requests for notifications
type NotificationController struct {
	notificationService *services.NotificationService
}

// NewNotificationController creates a new NotificationController
func NewNotificationController(notificationService *services.NotificationService) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

// GetNotifications handles GET /api/notifications
func (c *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get notifications
	notifications, total, err := c.notificationService.GetUserNotifications(userID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unread count
	unreadCount, _ := c.notificationService.GetUnreadCount(userID)

	// Convert to response format
	responses := make([]models.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = models.NotificationResponse{
			ID:                n.ID,
			UserID:            n.UserID,
			Type:              string(n.Type),
			Message:           n.Message,
			IsRead:            n.IsRead,
			RelatedEntityType: n.RelatedEntityType,
			RelatedEntityID:   n.RelatedEntityID,
			CreatedAt:         n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.NotificationListResponse{
		Success:     true,
		Data:        responses,
		UnreadCount: unreadCount,
		Page:        page,
		Limit:       limit,
		Total:       total,
	})
}

// MarkAsRead handles PATCH /api/notifications/:id/read
func (c *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	err = c.notificationService.MarkAsRead(notificationID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PATCH /api/notifications/read-all
func (c *NotificationController) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := c.notificationService.MarkAllAsRead(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "All notifications marked as read",
		"marked_count": count,
	})
}
