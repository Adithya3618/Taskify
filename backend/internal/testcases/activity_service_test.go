package testcases

import (
	"database/sql"
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to create a test database with activity logs
func newActivityTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create activity_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			user_name TEXT,
			action TEXT NOT NULL,
			entity_type TEXT,
			entity_id INTEGER,
			description TEXT,
			details TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create activity_logs table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_logs_project ON activity_logs(project_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

// Test ActivityService creation
func TestNewActivityService(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)
	if svc == nil {
		t.Error("NewActivityService() should not return nil")
	}
}

// Test ActivityRepository CreateActivityLog
func TestActivityRepository_CreateActivityLog(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	repo := repository.NewActivityRepository(db)

	activity := &models.ActivityLog{
		ProjectID:   1,
		UserID:      "user-123",
		UserName:    "Test User",
		Action:      models.ActivityTaskCreated,
		EntityType:  models.EntityTask,
		EntityID:    1,
		Description: "Created task 'Test Task'",
		Details:     "",
		CreatedAt:   time.Now(),
	}

	err := repo.CreateActivityLog(activity)
	if err != nil {
		t.Fatalf("CreateActivityLog() error = %v", err)
	}
}

// Test ActivityRepository GetActivityLogsByProject
func TestActivityRepository_GetActivityLogsByProject(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	repo := repository.NewActivityRepository(db)

	// Create multiple activity logs for project 1
	for i := 0; i < 5; i++ {
		activity := &models.ActivityLog{
			ProjectID:   1,
			UserID:      "user-123",
			UserName:    "Test User",
			Action:      models.ActivityTaskCreated,
			EntityType:  models.EntityTask,
			EntityID:    int64(i),
			Description: "Created task",
			CreatedAt:   time.Now(),
		}
		err := repo.CreateActivityLog(activity)
		if err != nil {
			t.Fatalf("CreateActivityLog() error = %v", err)
		}
	}

	// Create some logs for project 2
	for i := 0; i < 3; i++ {
		activity := &models.ActivityLog{
			ProjectID:   2,
			UserID:      "user-456",
			UserName:    "Other User",
			Action:      models.ActivityTaskCreated,
			EntityType:  models.EntityTask,
			EntityID:    int64(i),
			Description: "Created task",
			CreatedAt:   time.Now(),
		}
		err := repo.CreateActivityLog(activity)
		if err != nil {
			t.Fatalf("CreateActivityLog() error = %v", err)
		}
	}

	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}

	logs, total, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 5 {
		t.Errorf("GetActivityLogsByProject() len = %v, want 5", len(logs))
	}
	if total != 5 {
		t.Errorf("GetActivityLogsByProject() total = %v, want 5", total)
	}
}

// Test ActivityRepository GetActivityLogsByProject with pagination
func TestActivityRepository_GetActivityLogsByProject_Pagination(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	repo := repository.NewActivityRepository(db)

	// Create 10 activity logs
	for i := 0; i < 10; i++ {
		activity := &models.ActivityLog{
			ProjectID:   1,
			UserID:      "user-123",
			UserName:    "Test User",
			Action:      models.ActivityTaskCreated,
			EntityType:  models.EntityTask,
			EntityID:    int64(i),
			Description: "Created task",
			CreatedAt:   time.Now(),
		}
		err := repo.CreateActivityLog(activity)
		if err != nil {
			t.Fatalf("CreateActivityLog() error = %v", err)
		}
	}

	// Get page 1 with limit 3
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     3,
	}

	logs, total, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("GetActivityLogsByProject() page 1 len = %v, want 3", len(logs))
	}
	if total != 10 {
		t.Errorf("GetActivityLogsByProject() total = %v, want 10", total)
	}

	// Get page 2
	params.Page = 2
	logs, _, err = repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() page 2 error = %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("GetActivityLogsByProject() page 2 len = %v, want 3", len(logs))
	}

	// Get page 4 (should be empty)
	params.Page = 4
	logs, _, err = repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() page 4 error = %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("GetActivityLogsByProject() page 4 len = %v, want 0", len(logs))
	}
}

// Test ActivityRepository GetActivityLogsByProject with date range filter
func TestActivityRepository_GetActivityLogsByProject_DateFilter(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	repo := repository.NewActivityRepository(db)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	// Create activity logs with different dates
	activities := []struct {
		daysAgo int
	}{
		{0}, // today
		{0}, // today
		{1}, // yesterday
		{7}, // last week
	}

	for i, a := range activities {
		activity := &models.ActivityLog{
			ProjectID:   1,
			UserID:      "user-123",
			UserName:    "Test User",
			Action:      models.ActivityTaskCreated,
			EntityType:  models.EntityTask,
			EntityID:    int64(i),
			Description: "Created task",
			CreatedAt:   now.AddDate(0, 0, -a.daysAgo),
		}
		err := repo.CreateActivityLog(activity)
		if err != nil {
			t.Fatalf("CreateActivityLog() error = %v", err)
		}
	}

	// Filter from yesterday onwards
	from := yesterday
	params := repository.ActivityLogParams{
		ProjectID: 1,
		From:      &from,
		Page:      1,
		Limit:     10,
	}

	_, total, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if total != 3 {
		t.Errorf("GetActivityLogsByProject() with date filter total = %v, want 3", total)
	}
}

// Test ActivityService LogActivity
func TestActivityService_LogActivity(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)

	err := svc.LogActivity(
		1,
		"user-123",
		"Test User",
		models.ActivityTaskCreated,
		models.EntityTask,
		1,
		"Created task",
		"",
	)
	if err != nil {
		t.Fatalf("LogActivity() error = %v", err)
	}

	// Verify the log was created
	repo := repository.NewActivityRepository(db)
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 activity log, got %d", len(logs))
	}
	if logs[0].Action != models.ActivityTaskCreated {
		t.Errorf("Activity action = %v, want %v", logs[0].Action, models.ActivityTaskCreated)
	}
}

// Test ActivityService LogLabelCreated
func TestActivityService_LogLabelCreated(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)

	svc.LogLabelCreated(1, "user-123", "Test User", 1, "Test Label")

	// Verify the log was created
	repo := repository.NewActivityRepository(db)
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 activity log, got %d", len(logs))
	}
	if logs[0].Action != models.ActivityLabelCreated {
		t.Errorf("Activity action = %v, want %v", logs[0].Action, models.ActivityLabelCreated)
	}
}

// Test ActivityService LogLabelDeleted
func TestActivityService_LogLabelDeleted(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)

	svc.LogLabelDeleted(1, "user-123", "Test User", 1, "Test Label")

	// Verify the log was created
	repo := repository.NewActivityRepository(db)
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 activity log, got %d", len(logs))
	}
	if logs[0].Action != models.ActivityLabelDeleted {
		t.Errorf("Activity action = %v, want %v", logs[0].Action, models.ActivityLabelDeleted)
	}
}

// Test ActivityService LogLabelAssigned
func TestActivityService_LogLabelAssigned(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)

	svc.LogLabelAssigned(1, "user-123", "Test User", 1, "Test Label", "Test Task")

	// Verify the log was created
	repo := repository.NewActivityRepository(db)
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 activity log, got %d", len(logs))
	}
	if logs[0].Action != models.ActivityLabelAssigned {
		t.Errorf("Activity action = %v, want %v", logs[0].Action, models.ActivityLabelAssigned)
	}
}

// Test ActivityService LogLabelRemoved
func TestActivityService_LogLabelRemoved(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	svc := services.NewActivityService(db, pmService)

	svc.LogLabelRemoved(1, "user-123", "Test User", 1, "Test Label", "Test Task")

	// Verify the log was created
	repo := repository.NewActivityRepository(db)
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     10,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 activity log, got %d", len(logs))
	}
	if logs[0].Action != models.ActivityLabelRemoved {
		t.Errorf("Activity action = %v, want %v", logs[0].Action, models.ActivityLabelRemoved)
	}
}

// Test Concurrent Activity Logging (Event-Based pattern)
func TestActivityRepository_ConcurrentLogging(t *testing.T) {
	db := newActivityTestDB(t)
	defer db.Close()

	repo := repository.NewActivityRepository(db)

	// Simulate concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			activity := &models.ActivityLog{
				ProjectID:   1,
				UserID:      "user-123",
				UserName:    "Test User",
				Action:      models.ActivityTaskCreated,
				EntityType:  models.EntityTask,
				EntityID:    int64(idx),
				Description: "Created task",
				CreatedAt:   time.Now(),
			}
			err := repo.CreateActivityLog(activity)
			if err != nil {
				t.Errorf("CreateActivityLog() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all logs were created
	params := repository.ActivityLogParams{
		ProjectID: 1,
		Page:      1,
		Limit:     100,
	}
	logs, _, err := repo.GetActivityLogsByProject(params)
	if err != nil {
		t.Fatalf("GetActivityLogsByProject() error = %v", err)
	}

	if len(logs) != 10 {
		t.Errorf("Expected 10 activity logs after concurrent writes, got %d", len(logs))
	}
}
