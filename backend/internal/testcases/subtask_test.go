package testcases

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend/internal/controllers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// DB helper
// ---------------------------------------------------------------------------

func newSubtaskTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "subtask_test_*.db")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	schema := []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			role TEXT DEFAULT 'user',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_id TEXT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE stages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			position INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			stage_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			position INTEGER DEFAULT 0,
			deadline DATETIME,
			priority TEXT,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE subtasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			is_completed INTEGER DEFAULT 0,
			position INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX idx_subtasks_task ON subtasks(task_id)`,
		`CREATE INDEX idx_subtasks_task_position ON subtasks(task_id, position)`,
		`CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			author_name TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			t.Fatalf("db.Exec(schema) error = %v\nstmt: %s", err, stmt)
		}
	}

	return db
}

// seedSubtaskFixtures creates a project/stage/task owned by the given user and
// returns the taskID.
func seedSubtaskFixtures(t *testing.T, db *sql.DB, userID string) int64 {
	t.Helper()

	projRes, err := db.Exec(
		"INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)",
		userID, "Project", "Desc",
	)
	if err != nil {
		t.Fatalf("insert project error = %v", err)
	}
	projID, _ := projRes.LastInsertId()

	stageRes, err := db.Exec(
		"INSERT INTO stages (user_id, project_id, name, position) VALUES (?, ?, ?, ?)",
		userID, projID, "To Do", 0,
	)
	if err != nil {
		t.Fatalf("insert stage error = %v", err)
	}
	stageID, _ := stageRes.LastInsertId()

	taskRes, err := db.Exec(
		"INSERT INTO tasks (user_id, stage_id, title, description, position) VALUES (?, ?, ?, ?, ?)",
		userID, stageID, "Parent Task", "desc", 0,
	)
	if err != nil {
		t.Fatalf("insert task error = %v", err)
	}
	taskID, _ := taskRes.LastInsertId()

	return taskID
}

// ptr helpers
func strPtr(s string) *string  { return &s }
func boolPtr(b bool) *bool     { return &b }
func intPtr(i int) *int        { return &i }

// ---------------------------------------------------------------------------
// Service-layer tests
// ---------------------------------------------------------------------------

func TestSubtaskService_CreateSubtask_Success(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	subtask, err := svc.CreateSubtask(userID, taskID, "Buy milk", nil)
	if err != nil {
		t.Fatalf("CreateSubtask() error = %v", err)
	}

	if subtask.ID == 0 {
		t.Error("subtask ID should not be 0")
	}
	if subtask.TaskID != taskID {
		t.Errorf("TaskID = %d, want %d", subtask.TaskID, taskID)
	}
	if subtask.Title != "Buy milk" {
		t.Errorf("Title = %q, want %q", subtask.Title, "Buy milk")
	}
	if subtask.IsCompleted {
		t.Error("IsCompleted should be false on creation")
	}
	if subtask.Position != 0 {
		t.Errorf("Position = %d, want 0", subtask.Position)
	}
}

func TestSubtaskService_CreateSubtask_EmptyTitleRejected(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	_, err := svc.CreateSubtask(userID, taskID, "   ", nil)
	if err == nil {
		t.Fatal("CreateSubtask() error = nil, want ErrSubtaskTitleRequired")
	}
	if err != services.ErrSubtaskTitleRequired {
		t.Fatalf("CreateSubtask() error = %v, want %v", err, services.ErrSubtaskTitleRequired)
	}
}

func TestSubtaskService_CreateSubtask_InvalidTaskRejected(t *testing.T) {
	db := newSubtaskTestDB(t)
	svc := services.NewSubtaskService(db)

	_, err := svc.CreateSubtask("any-user", 99999, "Title", nil)
	if err == nil {
		t.Fatal("CreateSubtask() error = nil, want task not found error")
	}
	if err != services.ErrTaskNotFoundOrAccessDenied {
		t.Fatalf("CreateSubtask() error = %v, want %v", err, services.ErrTaskNotFoundOrAccessDenied)
	}
}

func TestSubtaskService_GetSubtasksByTask_OrderedByPosition(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	titles := []string{"First", "Second", "Third"}
	for _, title := range titles {
		if _, err := svc.CreateSubtask(userID, taskID, title, nil); err != nil {
			t.Fatalf("CreateSubtask(%q) error = %v", title, err)
		}
	}

	subtasks, err := svc.GetSubtasksByTask(userID, taskID)
	if err != nil {
		t.Fatalf("GetSubtasksByTask() error = %v", err)
	}
	if len(subtasks) != 3 {
		t.Fatalf("len(subtasks) = %d, want 3", len(subtasks))
	}
	for i, want := range titles {
		if subtasks[i].Title != want {
			t.Errorf("subtasks[%d].Title = %q, want %q", i, subtasks[i].Title, want)
		}
		if subtasks[i].Position != i {
			t.Errorf("subtasks[%d].Position = %d, want %d", i, subtasks[i].Position, i)
		}
	}
}

func TestSubtaskService_GetSubtasksByTask_EmptyWhenNone(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	subtasks, err := svc.GetSubtasksByTask(userID, taskID)
	if err != nil {
		t.Fatalf("GetSubtasksByTask() error = %v", err)
	}
	if len(subtasks) != 0 {
		t.Errorf("len(subtasks) = %d, want 0", len(subtasks))
	}
}

func TestSubtaskService_UpdateSubtask_Title(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	created, _ := svc.CreateSubtask(userID, taskID, "Original title", nil)

	updated, err := svc.UpdateSubtask(userID, created.ID, services.SubtaskPatch{
		Title: strPtr("New title"),
	})
	if err != nil {
		t.Fatalf("UpdateSubtask() error = %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("Title = %q, want %q", updated.Title, "New title")
	}
}

func TestSubtaskService_UpdateSubtask_ToggleCompletion(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	created, _ := svc.CreateSubtask(userID, taskID, "Do laundry", nil)
	if created.IsCompleted {
		t.Fatal("expected IsCompleted=false on creation")
	}

	completed, err := svc.UpdateSubtask(userID, created.ID, services.SubtaskPatch{
		IsCompleted: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("UpdateSubtask() error = %v", err)
	}
	if !completed.IsCompleted {
		t.Error("IsCompleted should be true after update")
	}

	uncompleted, err := svc.UpdateSubtask(userID, created.ID, services.SubtaskPatch{
		IsCompleted: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("UpdateSubtask() toggle back error = %v", err)
	}
	if uncompleted.IsCompleted {
		t.Error("IsCompleted should be false after toggling back")
	}
}

func TestSubtaskService_UpdateSubtask_EmptyTitleRejected(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	created, _ := svc.CreateSubtask(userID, taskID, "Valid", nil)

	_, err := svc.UpdateSubtask(userID, created.ID, services.SubtaskPatch{
		Title: strPtr(""),
	})
	if err != services.ErrSubtaskTitleRequired {
		t.Fatalf("UpdateSubtask() error = %v, want %v", err, services.ErrSubtaskTitleRequired)
	}
}

func TestSubtaskService_UpdateSubtask_NotFoundReturnsError(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	_, err := svc.UpdateSubtask(userID, 99999, services.SubtaskPatch{
		Title: strPtr("anything"),
	})
	if err == nil {
		t.Fatal("UpdateSubtask() error = nil, want access denied error")
	}
	if err != services.ErrSubtaskNotFoundOrAccessDenied {
		t.Fatalf("UpdateSubtask() error = %v, want %v", err, services.ErrSubtaskNotFoundOrAccessDenied)
	}
}

func TestSubtaskService_DeleteSubtask_Success(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	s1, _ := svc.CreateSubtask(userID, taskID, "First", nil)
	s2, _ := svc.CreateSubtask(userID, taskID, "Second", nil)
	_, _ = svc.CreateSubtask(userID, taskID, "Third", nil)
	_ = s2

	// Delete the middle one
	if err := svc.DeleteSubtask(userID, s1.ID); err != nil {
		t.Fatalf("DeleteSubtask() error = %v", err)
	}

	remaining, _ := svc.GetSubtasksByTask(userID, taskID)
	if len(remaining) != 2 {
		t.Fatalf("len(remaining) = %d, want 2", len(remaining))
	}
	// Positions should be compacted to 0, 1
	for i, st := range remaining {
		if st.Position != i {
			t.Errorf("remaining[%d].Position = %d, want %d", i, st.Position, i)
		}
	}
}

func TestSubtaskService_DeleteSubtask_NotFoundReturnsError(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	err := svc.DeleteSubtask(userID, 99999)
	if err == nil {
		t.Fatal("DeleteSubtask() error = nil, want access denied error")
	}
	if err != services.ErrSubtaskNotFoundOrAccessDenied {
		t.Fatalf("DeleteSubtask() error = %v, want %v", err, services.ErrSubtaskNotFoundOrAccessDenied)
	}
}

func TestSubtaskService_Reorder_MovingSubtask(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	s0, _ := svc.CreateSubtask(userID, taskID, "A", nil)
	s1, _ := svc.CreateSubtask(userID, taskID, "B", nil)
	s2, _ := svc.CreateSubtask(userID, taskID, "C", nil)
	_ = s1

	// Move "A" (position 0) to position 2 — should become last
	_, err := svc.UpdateSubtask(userID, s0.ID, services.SubtaskPatch{
		Position: intPtr(2),
	})
	if err != nil {
		t.Fatalf("UpdateSubtask(reorder) error = %v", err)
	}

	subtasks, _ := svc.GetSubtasksByTask(userID, taskID)
	if len(subtasks) != 3 {
		t.Fatalf("len = %d, want 3", len(subtasks))
	}
	// After moving "A" to last: order should be B, C, A
	want := []string{"B", "C", "A"}
	for i, st := range subtasks {
		if st.Title != want[i] {
			t.Errorf("subtasks[%d].Title = %q, want %q", i, st.Title, want[i])
		}
	}
	_ = s2
}

func TestSubtaskService_TitleTrimmed(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-1"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)

	subtask, err := svc.CreateSubtask(userID, taskID, "  trimmed title  ", nil)
	if err != nil {
		t.Fatalf("CreateSubtask() error = %v", err)
	}
	if subtask.Title != "trimmed title" {
		t.Errorf("Title = %q, want %q", subtask.Title, "trimmed title")
	}
}

// ---------------------------------------------------------------------------
// Controller-layer (HTTP) tests
// ---------------------------------------------------------------------------

func TestSubtaskController_CreateSubtask_Unauthorized(t *testing.T) {
	ctrl := controllers.NewSubtaskController(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/subtasks", nil)
	w := httptest.NewRecorder()

	ctrl.CreateSubtask(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("CreateSubtask() status = %d, want 401", w.Code)
	}
}

func TestSubtaskController_GetSubtasksByTask_Unauthorized(t *testing.T) {
	ctrl := controllers.NewSubtaskController(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/subtasks", nil)
	w := httptest.NewRecorder()

	ctrl.GetSubtasksByTask(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetSubtasksByTask() status = %d, want 401", w.Code)
	}
}

func TestSubtaskController_UpdateSubtask_Unauthorized(t *testing.T) {
	ctrl := controllers.NewSubtaskController(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/subtasks/1", nil)
	w := httptest.NewRecorder()

	ctrl.UpdateSubtask(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateSubtask() status = %d, want 401", w.Code)
	}
}

func TestSubtaskController_DeleteSubtask_Unauthorized(t *testing.T) {
	ctrl := controllers.NewSubtaskController(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/subtasks/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteSubtask(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteSubtask() status = %d, want 401", w.Code)
	}
}

func TestSubtaskController_CreateSubtask_EmptyTitleReturnsBadRequest(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-ctrl"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)
	ctrl := controllers.NewSubtaskController(svc)

	req := createRequestWithUser(http.MethodPost, "/api/tasks/1/subtasks", map[string]interface{}{
		"title": "",
	}, userID)
	req = mux.SetURLVars(req, map[string]string{"id": toString(taskID)})
	w := httptest.NewRecorder()

	ctrl.CreateSubtask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateSubtask() status = %d, want 400", w.Code)
	}
}

func TestSubtaskController_CRUD_RoundTrip(t *testing.T) {
	db := newSubtaskTestDB(t)
	const userID = "user-sub-ctrl"
	taskID := seedSubtaskFixtures(t, db, userID)
	svc := services.NewSubtaskService(db)
	ctrl := controllers.NewSubtaskController(svc)

	// Create
	createReq := createRequestWithUser(http.MethodPost, "/api/tasks/1/subtasks", map[string]interface{}{
		"title": "Write tests",
	}, userID)
	createReq = mux.SetURLVars(createReq, map[string]string{"id": toString(taskID)})
	createW := httptest.NewRecorder()
	ctrl.CreateSubtask(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("CreateSubtask() status = %d, want 201", createW.Code)
	}
	var created models.Subtask
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response error = %v", err)
	}
	if created.Title != "Write tests" {
		t.Errorf("Title = %q, want %q", created.Title, "Write tests")
	}
	if created.IsCompleted {
		t.Error("IsCompleted should be false on creation")
	}

	// List
	listReq := createRequestWithUser(http.MethodGet, "/api/tasks/1/subtasks", nil, userID)
	listReq = mux.SetURLVars(listReq, map[string]string{"id": toString(taskID)})
	listW := httptest.NewRecorder()
	ctrl.GetSubtasksByTask(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("GetSubtasksByTask() status = %d, want 200", listW.Code)
	}
	var listed []models.Subtask
	if err := json.NewDecoder(listW.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list response error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	// Update — mark complete
	trueVal := true
	updateReq := createRequestWithUser(http.MethodPatch, "/api/subtasks/1", map[string]interface{}{
		"is_completed": &trueVal,
	}, userID)
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": toString(created.ID)})
	updateW := httptest.NewRecorder()
	ctrl.UpdateSubtask(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("UpdateSubtask() status = %d, want 200", updateW.Code)
	}
	var updated models.Subtask
	if err := json.NewDecoder(updateW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update response error = %v", err)
	}
	if !updated.IsCompleted {
		t.Error("IsCompleted should be true after update")
	}

	// Delete
	deleteReq := createRequestWithUser(http.MethodDelete, "/api/subtasks/1", nil, userID)
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"id": toString(created.ID)})
	deleteW := httptest.NewRecorder()
	ctrl.DeleteSubtask(deleteW, deleteReq)

	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("DeleteSubtask() status = %d, want 204", deleteW.Code)
	}

	// Verify gone
	listReq2 := createRequestWithUser(http.MethodGet, "/api/tasks/1/subtasks", nil, userID)
	listReq2 = mux.SetURLVars(listReq2, map[string]string{"id": toString(taskID)})
	listW2 := httptest.NewRecorder()
	ctrl.GetSubtasksByTask(listW2, listReq2)

	var remaining []models.Subtask
	json.NewDecoder(listW2.Body).Decode(&remaining)
	if len(remaining) != 0 {
		t.Errorf("expected 0 subtasks after delete, got %d", len(remaining))
	}
}

func TestNewSubtaskController(t *testing.T) {
	ctrl := controllers.NewSubtaskController(nil)
	if ctrl == nil {
		t.Error("NewSubtaskController() should not return nil")
	}
}
