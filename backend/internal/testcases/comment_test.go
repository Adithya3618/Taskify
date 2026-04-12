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

func newCommentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "comment_test_*.db")
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
			user_id TEXT,
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
		`CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			author_name TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX idx_comments_task ON comments(task_id)`,
		`CREATE INDEX idx_comments_user ON comments(user_id)`,
		`CREATE TABLE subtasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			is_completed INTEGER DEFAULT 0,
			position INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			t.Fatalf("db.Exec(schema) error = %v", err)
		}
	}

	return db
}

// seedCommentFixtures inserts a user, project, stage and task.
// Returns (userID, taskID).
func seedCommentFixtures(t *testing.T, db *sql.DB) (string, int64) {
	t.Helper()

	const userID = "user-comment-1"
	if _, err := db.Exec(
		"INSERT INTO users (id, name, email) VALUES (?, ?, ?)",
		userID, "Alice", "alice@example.com",
	); err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	projRes, err := db.Exec(
		"INSERT INTO projects (user_id, name, description) VALUES (?, ?, ?)",
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
		userID, stageID, "Task", "desc", 0,
	)
	if err != nil {
		t.Fatalf("insert task error = %v", err)
	}
	taskID, _ := taskRes.LastInsertId()

	return userID, taskID
}

// ---------------------------------------------------------------------------
// Service-layer tests
// ---------------------------------------------------------------------------

func TestCommentService_CreateComment_Success(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	comment, err := svc.CreateComment(userID, taskID, "Hello world")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}

	if comment.ID == 0 {
		t.Error("comment ID should not be 0")
	}
	if comment.TaskID != taskID {
		t.Errorf("TaskID = %d, want %d", comment.TaskID, taskID)
	}
	if comment.UserID != userID {
		t.Errorf("UserID = %q, want %q", comment.UserID, userID)
	}
	if comment.Content != "Hello world" {
		t.Errorf("Content = %q, want %q", comment.Content, "Hello world")
	}
	if comment.AuthorName == "" {
		t.Error("AuthorName should not be empty")
	}
}

func TestCommentService_CreateComment_EmptyContentRejected(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	_, err := svc.CreateComment(userID, taskID, "   ")
	if err == nil {
		t.Fatal("CreateComment() error = nil, want ErrCommentContentRequired")
	}
	if err != services.ErrCommentContentRequired {
		t.Fatalf("CreateComment() error = %v, want %v", err, services.ErrCommentContentRequired)
	}
}

func TestCommentService_CreateComment_InvalidTaskRejected(t *testing.T) {
	db := newCommentTestDB(t)
	userID, _ := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	_, err := svc.CreateComment(userID, 99999, "Hello")
	if err == nil {
		t.Fatal("CreateComment() error = nil, want task not found error")
	}
	if err != services.ErrTaskNotFoundOrAccessDenied {
		t.Fatalf("CreateComment() error = %v, want %v", err, services.ErrTaskNotFoundOrAccessDenied)
	}
}

func TestCommentService_GetCommentsByTask_ReturnsAll(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	contents := []string{"First comment", "Second comment", "Third comment"}
	for _, c := range contents {
		if _, err := svc.CreateComment(userID, taskID, c); err != nil {
			t.Fatalf("CreateComment(%q) error = %v", c, err)
		}
	}

	comments, err := svc.GetCommentsByTask(userID, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask() error = %v", err)
	}
	if len(comments) != 3 {
		t.Fatalf("len(comments) = %d, want 3", len(comments))
	}
	for i, want := range contents {
		if comments[i].Content != want {
			t.Errorf("comments[%d].Content = %q, want %q", i, comments[i].Content, want)
		}
	}
}

func TestCommentService_GetCommentsByTask_EmptyWhenNone(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	comments, err := svc.GetCommentsByTask(userID, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask() error = %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("len(comments) = %d, want 0", len(comments))
	}
}

func TestCommentService_UpdateComment_Success(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	created, err := svc.CreateComment(userID, taskID, "Original")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}

	updated, err := svc.UpdateComment(userID, created.ID, "Updated content")
	if err != nil {
		t.Fatalf("UpdateComment() error = %v", err)
	}
	if updated.Content != "Updated content" {
		t.Errorf("Content = %q, want %q", updated.Content, "Updated content")
	}
	if updated.ID != created.ID {
		t.Errorf("ID = %d, want %d", updated.ID, created.ID)
	}
}

func TestCommentService_UpdateComment_EmptyContentRejected(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	created, _ := svc.CreateComment(userID, taskID, "Original")

	_, err := svc.UpdateComment(userID, created.ID, "")
	if err != services.ErrCommentContentRequired {
		t.Fatalf("UpdateComment() error = %v, want %v", err, services.ErrCommentContentRequired)
	}
}

func TestCommentService_UpdateComment_OtherUserCannotEdit(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	created, _ := svc.CreateComment(userID, taskID, "Owner comment")

	_, err := svc.UpdateComment("other-user", created.ID, "Hijacked")
	if err == nil {
		t.Fatal("UpdateComment() error = nil, want access denied error")
	}
	if err != services.ErrCommentNotFoundOrAccessDenied {
		t.Fatalf("UpdateComment() error = %v, want %v", err, services.ErrCommentNotFoundOrAccessDenied)
	}
}

func TestCommentService_DeleteComment_Success(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	created, _ := svc.CreateComment(userID, taskID, "To delete")

	if err := svc.DeleteComment(userID, created.ID); err != nil {
		t.Fatalf("DeleteComment() error = %v", err)
	}

	comments, _ := svc.GetCommentsByTask(userID, taskID)
	if len(comments) != 0 {
		t.Errorf("expected 0 comments after delete, got %d", len(comments))
	}
}

func TestCommentService_DeleteComment_OtherUserCannotDelete(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	created, _ := svc.CreateComment(userID, taskID, "Owner comment")

	err := svc.DeleteComment("other-user", created.ID)
	if err == nil {
		t.Fatal("DeleteComment() error = nil, want access denied error")
	}
	if err != services.ErrCommentNotFoundOrAccessDenied {
		t.Fatalf("DeleteComment() error = %v, want %v", err, services.ErrCommentNotFoundOrAccessDenied)
	}
}

func TestCommentService_ContentTrimmed(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)

	comment, err := svc.CreateComment(userID, taskID, "  trimmed content  ")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if comment.Content != "trimmed content" {
		t.Errorf("Content = %q, want %q", comment.Content, "trimmed content")
	}
}

// ---------------------------------------------------------------------------
// Controller-layer (HTTP) tests
// ---------------------------------------------------------------------------

func TestCommentController_CreateComment_Unauthorized(t *testing.T) {
	ctrl := controllers.NewCommentController(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/comments", nil)
	w := httptest.NewRecorder()

	ctrl.CreateComment(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("CreateComment() status = %d, want 401", w.Code)
	}
}

func TestCommentController_GetCommentsByTask_Unauthorized(t *testing.T) {
	ctrl := controllers.NewCommentController(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/comments", nil)
	w := httptest.NewRecorder()

	ctrl.GetCommentsByTask(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCommentsByTask() status = %d, want 401", w.Code)
	}
}

func TestCommentController_UpdateComment_Unauthorized(t *testing.T) {
	ctrl := controllers.NewCommentController(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/comments/1", nil)
	w := httptest.NewRecorder()

	ctrl.UpdateComment(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateComment() status = %d, want 401", w.Code)
	}
}

func TestCommentController_DeleteComment_Unauthorized(t *testing.T) {
	ctrl := controllers.NewCommentController(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/1", nil)
	w := httptest.NewRecorder()

	ctrl.DeleteComment(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteComment() status = %d, want 401", w.Code)
	}
}

func TestCommentController_CreateComment_EmptyContentReturnsBadRequest(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)
	ctrl := controllers.NewCommentController(svc)

	req := createRequestWithUser(http.MethodPost, "/api/tasks/1/comments", map[string]interface{}{
		"content": "",
	}, userID)
	req = mux.SetURLVars(req, map[string]string{"id": toString(taskID)})
	w := httptest.NewRecorder()

	ctrl.CreateComment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateComment() status = %d, want 400", w.Code)
	}
}

func TestCommentController_CRUD_RoundTrip(t *testing.T) {
	db := newCommentTestDB(t)
	userID, taskID := seedCommentFixtures(t, db)
	svc := services.NewCommentService(db)
	ctrl := controllers.NewCommentController(svc)

	// Create
	createReq := createRequestWithUser(http.MethodPost, "/api/tasks/1/comments", map[string]interface{}{
		"content": "My first comment",
	}, userID)
	createReq = mux.SetURLVars(createReq, map[string]string{"id": toString(taskID)})
	createW := httptest.NewRecorder()
	ctrl.CreateComment(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("CreateComment() status = %d, want 201", createW.Code)
	}
	var created models.Comment
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response error = %v", err)
	}
	if created.Content != "My first comment" {
		t.Errorf("Content = %q, want %q", created.Content, "My first comment")
	}

	// List
	listReq := createRequestWithUser(http.MethodGet, "/api/tasks/1/comments", nil, userID)
	listReq = mux.SetURLVars(listReq, map[string]string{"id": toString(taskID)})
	listW := httptest.NewRecorder()
	ctrl.GetCommentsByTask(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("GetCommentsByTask() status = %d, want 200", listW.Code)
	}
	var listed []models.Comment
	if err := json.NewDecoder(listW.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list response error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	// Update
	updateReq := createRequestWithUser(http.MethodPatch, "/api/comments/1", map[string]interface{}{
		"content": "Updated comment",
	}, userID)
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": toString(created.ID)})
	updateW := httptest.NewRecorder()
	ctrl.UpdateComment(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("UpdateComment() status = %d, want 200", updateW.Code)
	}
	var updated models.Comment
	if err := json.NewDecoder(updateW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update response error = %v", err)
	}
	if updated.Content != "Updated comment" {
		t.Errorf("Content = %q, want %q", updated.Content, "Updated comment")
	}

	// Delete
	deleteReq := createRequestWithUser(http.MethodDelete, "/api/comments/1", nil, userID)
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"id": toString(created.ID)})
	deleteW := httptest.NewRecorder()
	ctrl.DeleteComment(deleteW, deleteReq)

	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("DeleteComment() status = %d, want 204", deleteW.Code)
	}

	// Verify gone
	listReq2 := createRequestWithUser(http.MethodGet, "/api/tasks/1/comments", nil, userID)
	listReq2 = mux.SetURLVars(listReq2, map[string]string{"id": toString(taskID)})
	listW2 := httptest.NewRecorder()
	ctrl.GetCommentsByTask(listW2, listReq2)

	var remaining []models.Comment
	json.NewDecoder(listW2.Body).Decode(&remaining)
	if len(remaining) != 0 {
		t.Errorf("expected 0 comments after delete, got %d", len(remaining))
	}
}

func TestNewCommentController(t *testing.T) {
	ctrl := controllers.NewCommentController(nil)
	if ctrl == nil {
		t.Error("NewCommentController() should not return nil")
	}
}
