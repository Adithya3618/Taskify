package testcases

import (
	"database/sql"
	"testing"

	notifrepo "backend/internal/repository"
	"backend/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to create a test database with notification tables
func newNotificationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create notifications table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			is_read INTEGER DEFAULT 0,
			related_entity_type TEXT,
			related_entity_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create notifications table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

// Test NotificationService creation
func TestNewNotificationService(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	svc := services.NewNotificationService(db, nil)
	if svc == nil {
		t.Error("NewNotificationService() should not return nil")
	}
}

// Test NotificationRepository CreateNotification
func TestNotificationRepository_CreateNotification(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	notification, err := repo.CreateNotification(
		"user-123",
		"project_invite",
		"You were added to project 'Test Project'",
		"project",
		1,
	)
	if err != nil {
		t.Fatalf("CreateNotification() error = %v", err)
	}

	if notification.UserID != "user-123" {
		t.Errorf("CreateNotification() UserID = %v, want user-123", notification.UserID)
	}
	if notification.Type != "project_invite" {
		t.Errorf("CreateNotification() Type = %v, want project_invite", notification.Type)
	}
	if notification.Message != "You were added to project 'Test Project'" {
		t.Errorf("CreateNotification() Message = %v, want expected message", notification.Message)
	}
	if notification.IsRead {
		t.Error("CreateNotification() IsRead should be false")
	}
}

// Test NotificationRepository GetUserNotifications
func TestNotificationRepository_GetUserNotifications(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create multiple notifications for user
	for i := 0; i < 5; i++ {
		_, err := repo.CreateNotification(
			"user-123",
			"project_invite",
			"Notification "+string(rune('0'+i)),
			"project",
			int64(i),
		)
		if err != nil {
			t.Fatalf("CreateNotification() error = %v", err)
		}
	}

	// Get first page
	notifications, total, err := repo.GetUserNotifications("user-123", 1, 3)
	if err != nil {
		t.Fatalf("GetUserNotifications() error = %v", err)
	}
	if len(notifications) != 3 {
		t.Errorf("GetUserNotifications() len = %v, want 3", len(notifications))
	}
	if total != 5 {
		t.Errorf("GetUserNotifications() total = %v, want 5", total)
	}

	// Get second page
	notifications, _, err = repo.GetUserNotifications("user-123", 2, 3)
	if err != nil {
		t.Fatalf("GetUserNotifications() error = %v", err)
	}
	if len(notifications) != 2 {
		t.Errorf("GetUserNotifications() page 2 len = %v, want 2", len(notifications))
	}

	// Get notifications for non-existent user
	notifications, total, err = repo.GetUserNotifications("non-existent", 1, 10)
	if err != nil {
		t.Fatalf("GetUserNotifications() error = %v", err)
	}
	if len(notifications) != 0 {
		t.Errorf("GetUserNotifications() for non-existent len = %v, want 0", len(notifications))
	}
	if total != 0 {
		t.Errorf("GetUserNotifications() for non-existent total = %v, want 0", total)
	}
}

// Test NotificationRepository MarkAsRead
func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create a notification
	notification, err := repo.CreateNotification(
		"user-123",
		"task_assigned",
		"You were assigned a task",
		"task",
		1,
	)
	if err != nil {
		t.Fatalf("CreateNotification() error = %v", err)
	}

	// Mark as read
	err = repo.MarkAsRead(notification.ID, "user-123")
	if err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}

	// Verify it's marked as read
	notifications, _, err := repo.GetUserNotifications("user-123", 1, 10)
	if err != nil {
		t.Fatalf("GetUserNotifications() error = %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("GetUserNotifications() len = %v, want 1", len(notifications))
	}
	if !notifications[0].IsRead {
		t.Error("Notification should be marked as read")
	}
}

// Test NotificationRepository MarkAsRead - wrong user
func TestNotificationRepository_MarkAsRead_WrongUser(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create a notification
	notification, err := repo.CreateNotification(
		"user-123",
		"task_assigned",
		"You were assigned a task",
		"task",
		1,
	)
	if err != nil {
		t.Fatalf("CreateNotification() error = %v", err)
	}

	// Try to mark as read as different user
	err = repo.MarkAsRead(notification.ID, "wrong-user")
	if err == nil {
		t.Error("MarkAsRead() should return error for wrong user")
	}
}

// Test NotificationRepository MarkAllAsRead
func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create multiple unread notifications
	for i := 0; i < 3; i++ {
		_, err := repo.CreateNotification(
			"user-123",
			"project_invite",
			"Notification "+string(rune('0'+i)),
			"project",
			int64(i),
		)
		if err != nil {
			t.Fatalf("CreateNotification() error = %v", err)
		}
	}

	// Mark all as read
	count, err := repo.MarkAllAsRead("user-123")
	if err != nil {
		t.Fatalf("MarkAllAsRead() error = %v", err)
	}
	if count != 3 {
		t.Errorf("MarkAllAsRead() count = %v, want 3", count)
	}

	// Verify all are marked as read
	notifications, _, err := repo.GetUserNotifications("user-123", 1, 10)
	if err != nil {
		t.Fatalf("GetUserNotifications() error = %v", err)
	}
	for _, n := range notifications {
		if !n.IsRead {
			t.Error("All notifications should be marked as read")
		}
	}
}

// Test NotificationRepository GetUnreadCount
func TestNotificationRepository_GetUnreadCount(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create 5 notifications
	for i := 0; i < 5; i++ {
		_, err := repo.CreateNotification(
			"user-123",
			"project_invite",
			"Notification",
			"project",
			int64(i),
		)
		if err != nil {
			t.Fatalf("CreateNotification() error = %v", err)
		}
	}

	// Check unread count
	count, err := repo.GetUnreadCount("user-123")
	if err != nil {
		t.Fatalf("GetUnreadCount() error = %v", err)
	}
	if count != 5 {
		t.Errorf("GetUnreadCount() = %v, want 5", count)
	}

	// Mark 2 as read
	notifications, _, _ := repo.GetUserNotifications("user-123", 1, 10)
	err = repo.MarkAsRead(notifications[0].ID, "user-123")
	if err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	err = repo.MarkAsRead(notifications[1].ID, "user-123")
	if err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}

	// Check unread count again
	count, err = repo.GetUnreadCount("user-123")
	if err != nil {
		t.Fatalf("GetUnreadCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("GetUnreadCount() after marking 2 read = %v, want 3", count)
	}
}

// Test Duplicate Deadline Notification Prevention
func TestNotificationRepository_CheckDuplicateDeadlineNotification(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// First check should return false (no duplicate)
	exists, err := repo.CheckDuplicateDeadlineNotification("user-123", 1)
	if err != nil {
		t.Fatalf("CheckDuplicateDeadlineNotification() error = %v", err)
	}
	if exists {
		t.Error("CheckDuplicateDeadlineNotification() should return false for first check")
	}

	// Create a deadline notification
	_, err = repo.CreateNotification("user-123", "deadline_near", "Task due soon", "task", 1)
	if err != nil {
		t.Fatalf("CreateNotification() error = %v", err)
	}

	// Second check should return true (duplicate exists)
	exists, err = repo.CheckDuplicateDeadlineNotification("user-123", 1)
	if err != nil {
		t.Fatalf("CheckDuplicateDeadlineNotification() error = %v", err)
	}
	if !exists {
		t.Error("CheckDuplicateDeadlineNotification() should return true for duplicate")
	}
}

// Test NotificationRepository GetNotificationByID
func TestNotificationRepository_GetNotificationByID(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Create a notification
	created, err := repo.CreateNotification(
		"user-123",
		"task_assigned",
		"You were assigned a task",
		"task",
		1,
	)
	if err != nil {
		t.Fatalf("CreateNotification() error = %v", err)
	}

	// Get notification by ID
	fetched, err := repo.GetNotificationByID(created.ID)
	if err != nil {
		t.Fatalf("GetNotificationByID() error = %v", err)
	}
	if fetched == nil {
		t.Fatal("GetNotificationByID() returned nil")
	}
	if fetched.ID != created.ID {
		t.Errorf("GetNotificationByID() id = %v, want %v", fetched.ID, created.ID)
	}
}

// Test NotificationRepository GetNotificationByID - not found
func TestNotificationRepository_GetNotificationByID_NotFound(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	repo := notifrepo.NewNotificationRepository(db)

	// Get notification by non-existent ID
	fetched, err := repo.GetNotificationByID(999)
	if err != nil {
		t.Fatalf("GetNotificationByID() error = %v", err)
	}
	if fetched != nil {
		t.Error("GetNotificationByID() should return nil for non-existent ID")
	}
}

// Test NotificationService NotifyMemberAdded - self notification
func TestNotificationService_NotifyMemberAdded_SelfNotification(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	svc := services.NewNotificationService(db, nil)

	// Should not notify actor themselves
	err := svc.NotifyMemberAdded(1, "user-1", "user-1", "Test User", "Test Project")
	if err != nil {
		t.Fatalf("NotifyMemberAdded() error = %v", err)
	}

	// Verify no notification was created
	repo := notifrepo.NewNotificationRepository(db)
	count, err := repo.GetUnreadCount("user-1")
	if err != nil {
		t.Fatalf("GetUnreadCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 notifications for self-notification, got %d", count)
	}
}

// Test NotificationService NotifyTaskAssigned - self notification
func TestNotificationService_NotifyTaskAssigned_SelfNotification(t *testing.T) {
	db := newNotificationTestDB(t)
	defer db.Close()

	svc := services.NewNotificationService(db, nil)

	// Should not notify actor themselves
	err := svc.NotifyTaskAssigned(1, "user-1", "user-1", "Test User", "Test Task")
	if err != nil {
		t.Fatalf("NotifyTaskAssigned() error = %v", err)
	}

	// Verify no notification was created
	repo := notifrepo.NewNotificationRepository(db)
	count, err := repo.GetUnreadCount("user-1")
	if err != nil {
		t.Fatalf("GetUnreadCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 notifications for self-notification, got %d", count)
	}
}
