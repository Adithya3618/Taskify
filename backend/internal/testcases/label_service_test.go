package testcases

import (
	"database/sql"
	"testing"

	"backend/internal/repository"
	"backend/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to create a test database with label tables
func newLabelTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create labels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#808080',
			created_by TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create labels table: %v", err)
	}

	// Create task_labels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_labels (
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (task_id, label_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create task_labels table: %v", err)
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

	return db
}

// Test LabelService creation
func TestNewLabelService(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	activityService := services.NewActivityService(db, pmService)

	svc := services.NewLabelService(db, pmService, activityService)
	if svc == nil {
		t.Error("NewLabelService() should not return nil")
	}
}

// Test LabelRepository CreateLabel
func TestLabelRepository_CreateLabel_Success(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	labelRepo := repository.NewLabelRepository(db)
	created, err := labelRepo.CreateLabel(1, "Test Label", "#FF5733", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	if created.Name != "Test Label" {
		t.Errorf("CreateLabel() name = %v, want Test Label", created.Name)
	}
	if created.Color != "#FF5733" {
		t.Errorf("CreateLabel() color = %v, want #FF5733", created.Color)
	}
}

// Test LabelRepository CreateLabel with default color
func TestLabelRepository_CreateLabel_DefaultColor(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	repo := repository.NewLabelRepository(db)
	created, err := repo.CreateLabel(1, "No Color Label", "", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	if created.Color != "#808080" {
		t.Errorf("CreateLabel() color = %v, want #808080 (default)", created.Color)
	}
}

// Test LabelRepository CreateLabel duplicate name
func TestLabelRepository_CreateLabel_DuplicateName(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	repo := repository.NewLabelRepository(db)

	// Create first label
	_, err := repo.CreateLabel(1, "Duplicate Label", "#FF0000", "user-1")
	if err != nil {
		t.Fatalf("First CreateLabel() error = %v", err)
	}

	// Try to create duplicate
	_, err = repo.CreateLabel(1, "Duplicate Label", "#00FF00", "user-1")
	if err == nil {
		t.Error("CreateLabel() should return error for duplicate name")
	}
}

// Test LabelRepository GetLabelByID
func TestLabelRepository_GetLabelByID(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	repo := repository.NewLabelRepository(db)

	// Create a label
	created, err := repo.CreateLabel(1, "Get Test Label", "#00FF00", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	// Get the label
	fetched, err := repo.GetLabelByID(created.ID)
	if err != nil {
		t.Fatalf("GetLabelByID() error = %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("GetLabelByID() id = %v, want %v", fetched.ID, created.ID)
	}
	if fetched.Name != "Get Test Label" {
		t.Errorf("GetLabelByID() name = %v, want Get Test Label", fetched.Name)
	}
}

// Test LabelRepository GetLabelsByProject
func TestLabelRepository_GetLabelsByProject(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	repo := repository.NewLabelRepository(db)

	// Create multiple labels for project 1
	_, err := repo.CreateLabel(1, "Label 1", "#FF0000", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}
	_, err = repo.CreateLabel(1, "Label 2", "#00FF00", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}
	_, err = repo.CreateLabel(1, "Label 3", "#0000FF", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	// Get labels for project 1
	labels, err := repo.GetLabelsByProject(1)
	if err != nil {
		t.Fatalf("GetLabelsByProject() error = %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("GetLabelsByProject() len = %v, want 3", len(labels))
	}

	// Get labels for project 2 (should be empty)
	labels, err = repo.GetLabelsByProject(2)
	if err != nil {
		t.Fatalf("GetLabelsByProject() error = %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("GetLabelsByProject() len = %v, want 0 for empty project", len(labels))
	}
}

// Test LabelRepository DeleteLabel
func TestLabelRepository_DeleteLabel(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	repo := repository.NewLabelRepository(db)

	// Create a label
	created, err := repo.CreateLabel(1, "Delete Test Label", "#FF0000", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	// Delete the label
	err = repo.DeleteLabel(created.ID)
	if err != nil {
		t.Fatalf("DeleteLabel() error = %v", err)
	}

	// Try to get the deleted label
	_, err = repo.GetLabelByID(created.ID)
	if err != sql.ErrNoRows {
		t.Errorf("GetLabelByID() after delete should return ErrNoRows, got %v", err)
	}
}

// Test TaskLabelRepository
func TestTaskLabelRepository(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	taskLabelRepo := repository.NewTaskLabelRepository(db)
	labelRepo := repository.NewLabelRepository(db)

	// Create a label
	label, err := labelRepo.CreateLabel(1, "Task Label Test", "#FF0000", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	// Assign label to task
	err = taskLabelRepo.AssignLabelToTask(1, label.ID)
	if err != nil {
		t.Fatalf("AssignLabelToTask() error = %v", err)
	}

	// Get task labels
	labels, err := taskLabelRepo.GetLabelsByTask(1)
	if err != nil {
		t.Fatalf("GetLabelsByTask() error = %v", err)
	}
	if len(labels) != 1 {
		t.Errorf("GetLabelsByTask() len = %v, want 1", len(labels))
	}

	// Remove label from task
	err = taskLabelRepo.RemoveLabelFromTask(1, label.ID)
	if err != nil {
		t.Fatalf("RemoveLabelFromTask() error = %v", err)
	}

	// Check label is removed
	labels, err = taskLabelRepo.GetLabelsByTask(1)
	if err != nil {
		t.Fatalf("GetLabelsByTask() error = %v", err)
	}
	if len(labels) != 0 {
		t.Error("GetLabelsByTask() should return 0 after removal")
	}
}

// Test TaskLabelRepository GetTaskIDsByLabel
func TestTaskLabelRepository_GetTaskIDsByLabel(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	taskLabelRepo := repository.NewTaskLabelRepository(db)
	labelRepo := repository.NewLabelRepository(db)

	// Create a label
	label, err := labelRepo.CreateLabel(1, "Get Tasks Label", "#FF0000", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}

	// Assign label to multiple tasks
	err = taskLabelRepo.AssignLabelToTask(1, label.ID)
	if err != nil {
		t.Fatalf("AssignLabelToTask() error = %v", err)
	}
	err = taskLabelRepo.AssignLabelToTask(2, label.ID)
	if err != nil {
		t.Fatalf("AssignLabelToTask() error = %v", err)
	}
	err = taskLabelRepo.AssignLabelToTask(3, label.ID)
	if err != nil {
		t.Fatalf("AssignLabelToTask() error = %v", err)
	}

	// Get task IDs by label
	taskIDs, err := taskLabelRepo.GetTaskIDsByLabel(label.ID)
	if err != nil {
		t.Fatalf("GetTaskIDsByLabel() error = %v", err)
	}

	if len(taskIDs) != 3 {
		t.Errorf("GetTaskIDsByLabel() len = %v, want 3", len(taskIDs))
	}
}

// Test Label validation for invalid color format
func TestLabelRepository_InvalidColorFormat(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	// Test that invalid color formats are stored (validation happens at service level)
	repo := repository.NewLabelRepository(db)
	label, err := repo.CreateLabel(1, "Test Label", "invalid", "user-1")
	if err != nil {
		t.Fatalf("CreateLabel() error = %v", err)
	}
	if label.Color != "invalid" {
		t.Errorf("CreateLabel() stored invalid color = %v", label.Color)
	}
}

// Test Label creation without project access
func TestLabelService_NoProjectAccess(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	pmService := services.NewProjectMemberService(db)
	activityService := services.NewActivityService(db, pmService)
	svc := services.NewLabelService(db, pmService, activityService)

	// Try to create label without project membership
	_, err := svc.CreateLabel(999, "Orphan Label", "#FF0000", "user-1", "Test User")
	if err == nil {
		t.Error("CreateLabel() should return error for non-existent project")
	}
}

// Test GetLabelsByTask for task with no labels
func TestTaskLabelRepository_NoLabels(t *testing.T) {
	db := newLabelTestDB(t)
	defer db.Close()

	taskLabelRepo := repository.NewTaskLabelRepository(db)

	// Get labels for task with no labels
	labels, err := taskLabelRepo.GetLabelsByTask(999)
	if err != nil {
		t.Fatalf("GetLabelsByTask() error = %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("GetLabelsByTask() for task with no labels len = %v, want 0", len(labels))
	}
}
