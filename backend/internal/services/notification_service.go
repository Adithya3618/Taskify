package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"backend/internal/auth/repository"
	"backend/internal/auth/services"
	"backend/internal/models"
	notifrepo "backend/internal/repository"
)

// NotificationService handles business logic for notifications
type NotificationService struct {
	db           *sql.DB
	notifRepo    *notifrepo.NotificationRepository
	userRepo     *repository.UserRepository
	emailService *services.EmailService
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(db *sql.DB, emailService *services.EmailService) *NotificationService {
	return &NotificationService{
		db:           db,
		notifRepo:    notifrepo.NewNotificationRepository(db),
		userRepo:     repository.NewUserRepository(db),
		emailService: emailService,
	}
}

// NotifyMemberAdded sends a notification when a user is added to a project
func (s *NotificationService) NotifyMemberAdded(projectID int64, addedUserID, actorID, actorName, projectName string) error {
	// Don't notify the actor themselves
	if addedUserID == actorID {
		return nil
	}

	message := fmt.Sprintf("You were added to project '%s' by %s", projectName, actorName)

	// Create notification
	_, err := s.notifRepo.CreateNotification(
		addedUserID,
		models.NotificationMemberAdded,
		message,
		string(models.EntityProject),
		projectID,
	)
	if err != nil {
		log.Printf("Failed to create member added notification: %v", err)
		return err
	}

	// Send email asynchronously (non-blocking)
	go s.sendEmailNotification(addedUserID, "Project Invitation", message)

	return nil
}

// NotifyTaskAssigned sends a notification when a task is assigned to a user
func (s *NotificationService) NotifyTaskAssigned(taskID int64, assignedUserID, actorID, actorName, taskTitle string) error {
	// Don't notify the actor themselves
	if assignedUserID == actorID {
		return nil
	}

	message := fmt.Sprintf("You were assigned task '%s' by %s", taskTitle, actorName)

	// Create notification
	_, err := s.notifRepo.CreateNotification(
		assignedUserID,
		models.NotificationTaskAssigned,
		message,
		string(models.EntityTask),
		taskID,
	)
	if err != nil {
		log.Printf("Failed to create task assigned notification: %v", err)
		return err
	}

	// Send email asynchronously (non-blocking)
	go s.sendEmailNotification(assignedUserID, "Task Assigned", message)

	return nil
}

// NotifyDeadlineNear sends a notification when a task deadline is approaching
func (s *NotificationService) NotifyDeadlineNear(taskID int64, userID, taskTitle string) error {
	// Check for duplicate notification
	exists, err := s.notifRepo.CheckDuplicateDeadlineNotification(userID, taskID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already notified
	}

	message := fmt.Sprintf("Task '%s' is due within 24 hours", taskTitle)

	// Create notification
	_, err = s.notifRepo.CreateNotification(
		userID,
		models.NotificationDeadlineNear,
		message,
		string(models.EntityTask),
		taskID,
	)
	if err != nil {
		log.Printf("Failed to create deadline notification: %v", err)
		return err
	}

	// Send email asynchronously (non-blocking)
	go s.sendEmailNotification(userID, "Deadline Reminder", message)

	return nil
}

// GetUserNotifications retrieves paginated notifications for a user
func (s *NotificationService) GetUserNotifications(userID string, page, limit int) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.notifRepo.GetUserNotifications(userID, page, limit)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID int64, userID string) error {
	return s.notifRepo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID string) (int64, error) {
	return s.notifRepo.MarkAllAsRead(userID)
}

// GetUnreadCount returns the count of unread notifications
func (s *NotificationService) GetUnreadCount(userID string) (int64, error) {
	return s.notifRepo.GetUnreadCount(userID)
}

// sendEmailNotification sends an email notification (non-blocking)
func (s *NotificationService) sendEmailNotification(userID, subject, message string) {
	if s.emailService == nil {
		return
	}

	// Get user email
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		log.Printf("Could not find user %s for email notification: %v", userID, err)
		return
	}

	email := user.Email
	if email == "" {
		log.Printf("User %s has no email for notification", userID)
		return
	}

	// Send email - the service handles the formatting
	err = s.emailService.SendNotification(email, subject, message)
	if err != nil {
		log.Printf("Failed to send email notification to %s: %v", email, err)
	}
}

// CheckDeadlines is a background job that checks for approaching deadlines
// It should be called periodically (e.g., every 15 minutes)
func (s *NotificationService) CheckDeadlines() error {
	// Find tasks with deadlines in the next 24 hours that haven't been notified
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.assigned_to
		FROM tasks t
		WHERE t.deadline IS NOT NULL
		AND t.deadline > datetime('now')
		AND t.deadline <= datetime('now', '+24 hours')
		AND t.assigned_to IS NOT NULL
		AND t.assigned_to != ''
	`)
	if err != nil {
		return fmt.Errorf("failed to query deadline tasks: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskID int64
		var taskTitle, assignedTo string

		if err := rows.Scan(&taskID, &taskTitle, &assignedTo); err != nil {
			log.Printf("Failed to scan task for deadline check: %v", err)
			continue
		}

		// Send notification
		if err := s.NotifyDeadlineNear(taskID, assignedTo, taskTitle); err != nil {
			log.Printf("Failed to notify deadline for task %d: %v", taskID, err)
		}
	}

	return nil
}

// StartDeadlineChecker starts a background goroutine that checks deadlines periodically
func (s *NotificationService) StartDeadlineChecker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.CheckDeadlines(); err != nil {
				log.Printf("Deadline check error: %v", err)
			}
		}
	}()
}
