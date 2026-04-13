package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(userID string, notificationType models.NotificationType, message, entityType string, entityID int64) (*models.Notification, error) {
	result, err := r.db.Exec(
		`INSERT INTO notifications (user_id, type, message, related_entity_type, related_entity_id, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID, notificationType, message, entityType, entityID, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return &models.Notification{
		ID:                id,
		UserID:            userID,
		Type:              notificationType,
		Message:           message,
		IsRead:            false,
		RelatedEntityType: entityType,
		RelatedEntityID:   entityID,
		CreatedAt:         time.Now(),
	}, nil
}

// GetUserNotifications retrieves paginated notifications for a user
func (r *NotificationRepository) GetUserNotifications(userID string, page, limit int) ([]models.Notification, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %v", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	rows, err := r.db.Query(`
		SELECT id, user_id, type, message, is_read, related_entity_type, related_entity_id, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %v", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var isRead int
		var entityType sql.NullString
		var entityID sql.NullInt64

		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &isRead, &entityType, &entityID, &n.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %v", err)
		}

		n.IsRead = isRead == 1
		if entityType.Valid {
			n.RelatedEntityType = entityType.String
		}
		if entityID.Valid {
			n.RelatedEntityID = entityID.Int64
		}

		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read (only if owned by user)
func (r *NotificationRepository) MarkAsRead(notificationID int64, userID string) error {
	result, err := r.db.Exec(
		"UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?",
		notificationID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or access denied")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(userID string) (int64, error) {
	result, err := r.db.Exec(
		"UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0",
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0",
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %v", err)
	}
	return count, nil
}

// CheckDuplicateDeadlineNotification checks if a deadline notification already exists for this task/user
func (r *NotificationRepository) CheckDuplicateDeadlineNotification(userID string, taskID int64) (bool, error) {
	var count int64
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM notifications 
		WHERE user_id = ? AND type = ? AND related_entity_id = ?
	`, userID, models.NotificationDeadlineNear, taskID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate notification: %v", err)
	}
	return count > 0, nil
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(notificationID int64) (*models.Notification, error) {
	var n models.Notification
	var isRead int
	var entityType sql.NullString
	var entityID sql.NullInt64

	err := r.db.QueryRow(`
		SELECT id, user_id, type, message, is_read, related_entity_type, related_entity_id, created_at
		FROM notifications
		WHERE id = ?
	`, notificationID).Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &isRead, &entityType, &entityID, &n.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %v", err)
	}

	n.IsRead = isRead == 1
	if entityType.Valid {
		n.RelatedEntityType = entityType.String
	}
	if entityID.Valid {
		n.RelatedEntityID = entityID.Int64
	}

	return &n, nil
}
