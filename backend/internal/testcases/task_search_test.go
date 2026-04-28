package testcases

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"backend/internal/controllers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func TestTaskService_SearchProjectTasksMatchesTitleAndDescription(t *testing.T) {
	db := newTaskSearchTestDB(t)
	defer db.Close()

	projectID, todoStageID, doneStageID := seedTaskSearchProject(t, db, "user-1")
	otherProjectID, otherStageID, _ := seedTaskSearchProject(t, db, "user-2")
	highPriority := "high"
	deadline := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	assignedTo := "user-2"

	titleMatchID := seedTaskSearchTask(t, db, "user-1", todoStageID, "Fix login search bug", "OAuth callback", &deadline, &highPriority, &assignedTo)
	descriptionMatchID := seedTaskSearchTask(t, db, "user-1", doneStageID, "Dashboard polish", "Improve Search results panel", nil, nil, nil)
	seedTaskSearchTask(t, db, "user-1", todoStageID, "Unrelated task", "No match here", nil, nil, nil)
	seedTaskSearchTask(t, db, "user-2", otherStageID, "Search in other project", "Must not leak", nil, nil, nil)

	service := services.NewTaskService(db, nil)
	results, err := service.SearchProjectTasks("user-1", projectID, "search")
	if err != nil {
		t.Fatalf("SearchProjectTasks() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("SearchProjectTasks() len = %d, want 2: %+v", len(results), results)
	}
	if results[0].TaskID != titleMatchID {
		t.Fatalf("first TaskID = %d, want %d", results[0].TaskID, titleMatchID)
	}
	if results[0].StageName != "To Do" {
		t.Fatalf("first StageName = %q, want To Do", results[0].StageName)
	}
	if results[0].Priority == nil || *results[0].Priority != "high" {
		t.Fatalf("first Priority = %v, want high", results[0].Priority)
	}
	if results[0].Deadline == nil || !results[0].Deadline.Equal(deadline) {
		t.Fatalf("first Deadline = %v, want %v", results[0].Deadline, deadline)
	}
	if results[0].AssignedTo == nil || *results[0].AssignedTo != "user-2" {
		t.Fatalf("first AssignedTo = %v, want user-2", results[0].AssignedTo)
	}

	if results[1].TaskID != descriptionMatchID {
		t.Fatalf("second TaskID = %d, want %d", results[1].TaskID, descriptionMatchID)
	}
	if results[1].StageName != "Done" {
		t.Fatalf("second StageName = %q, want Done", results[1].StageName)
	}

	otherResults, err := service.SearchProjectTasks("user-2", otherProjectID, "search")
	if err != nil {
		t.Fatalf("SearchProjectTasks(other project) error = %v", err)
	}
	if len(otherResults) != 1 || otherResults[0].StageID != otherStageID {
		t.Fatalf("other project results = %+v, want only other project task", otherResults)
	}
}

func TestTaskService_SearchProjectTasksReturnsEmptyArray(t *testing.T) {
	db := newTaskSearchTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTaskSearchProject(t, db, "user-1")
	seedTaskSearchTask(t, db, "user-1", stageID, "Build cards", "Kanban work", nil, nil, nil)

	service := services.NewTaskService(db, nil)
	results, err := service.SearchProjectTasks("user-1", projectID, "missing")
	if err != nil {
		t.Fatalf("SearchProjectTasks() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("SearchProjectTasks() len = %d, want 0", len(results))
	}
}

func TestTaskController_SearchProjectTasksReturnsPlainArray(t *testing.T) {
	db := newTaskSearchTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTaskSearchProject(t, db, "user-1")
	taskID := seedTaskSearchTask(t, db, "user-1", stageID, "Searchable backend task", "Endpoint work", nil, nil, nil)

	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/tasks/search?q=backend", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.SearchProjectTasks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SearchProjectTasks() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var results []models.TaskSearchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].TaskID != taskID {
		t.Fatalf("result TaskID = %d, want %d", results[0].TaskID, taskID)
	}
}

func TestTaskService_SearchProjectTasksValidatesProjectAndAccess(t *testing.T) {
	db := newTaskSearchTestDB(t)
	defer db.Close()

	projectID, _, _ := seedTaskSearchProject(t, db, "user-1")
	service := services.NewTaskService(db, nil)

	tests := []struct {
		name      string
		userID    string
		projectID int64
		query     string
		wantCode  string
	}{
		{
			name:      "blank query",
			userID:    "user-1",
			projectID: projectID,
			query:     "   ",
			wantCode:  "INVALID_REQUEST",
		},
		{
			name:      "missing project",
			userID:    "user-1",
			projectID: projectID + 100,
			query:     "task",
			wantCode:  "PROJECT_NOT_FOUND",
		},
		{
			name:      "no project access",
			userID:    "user-2",
			projectID: projectID,
			query:     "task",
			wantCode:  "ACCESS_DENIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.SearchProjectTasks(tt.userID, tt.projectID, tt.query)
			if err == nil {
				t.Fatalf("SearchProjectTasks() error = nil, want %s with results %+v", tt.wantCode, results)
			}
			serviceErr, ok := services.IsServiceError(err)
			if !ok {
				t.Fatalf("SearchProjectTasks() error = %T %v, want ServiceError", err, err)
			}
			if serviceErr.Code != tt.wantCode {
				t.Fatalf("ServiceError.Code = %q, want %q", serviceErr.Code, tt.wantCode)
			}
		})
	}
}

func TestTaskController_SearchProjectTasksValidatesRequest(t *testing.T) {
	db := newTaskSearchTestDB(t)
	defer db.Close()

	projectID, _, _ := seedTaskSearchProject(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)

	tests := []struct {
		name       string
		url        string
		userID     string
		projectVar string
		wantStatus int
	}{
		{
			name:       "missing query",
			url:        "/api/projects/1/tasks/search",
			userID:     "user-1",
			projectVar: toString(projectID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid project id",
			url:        "/api/projects/nope/tasks/search?q=task",
			userID:     "user-1",
			projectVar: "nope",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing project",
			url:        "/api/projects/999/tasks/search?q=task",
			userID:     "user-1",
			projectVar: toString(projectID + 100),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no project access",
			url:        "/api/projects/1/tasks/search?q=task",
			userID:     "user-2",
			projectVar: toString(projectID),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unauthorized",
			url:        "/api/projects/1/tasks/search?q=task",
			userID:     "",
			projectVar: toString(projectID),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithUser(http.MethodGet, tt.url, nil, tt.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.projectVar})
			w := httptest.NewRecorder()

			controller.SearchProjectTasks(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("SearchProjectTasks() status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func newTaskSearchTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "task_search_*.db")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	schema := []string{
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_id TEXT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE project_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			invited_by TEXT,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, user_id)
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
			start_date DATETIME,
			deadline DATETIME,
			priority TEXT,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, statement := range schema {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			t.Fatalf("db.Exec(schema) error = %v", err)
		}
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func seedTaskSearchProject(t *testing.T, db *sql.DB, ownerID string) (int64, int64, int64) {
	t.Helper()

	result, err := db.Exec(
		"INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)",
		ownerID, "Project", "Description",
	)
	if err != nil {
		t.Fatalf("insert project error = %v", err)
	}
	projectID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("project LastInsertId() error = %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO project_members (project_id, user_id, role, invited_by) VALUES (?, ?, ?, ?)",
		projectID, ownerID, "owner", ownerID,
	); err != nil {
		t.Fatalf("insert project member error = %v", err)
	}

	todoStageID := seedTaskSearchStage(t, db, ownerID, projectID, "To Do", 0)
	doneStageID := seedTaskSearchStage(t, db, ownerID, projectID, "Done", 1)
	return projectID, todoStageID, doneStageID
}

func seedTaskSearchStage(t *testing.T, db *sql.DB, userID string, projectID int64, name string, position int) int64 {
	t.Helper()

	result, err := db.Exec(
		"INSERT INTO stages (user_id, project_id, name, position) VALUES (?, ?, ?, ?)",
		userID, projectID, name, position,
	)
	if err != nil {
		t.Fatalf("insert stage error = %v", err)
	}
	stageID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("stage LastInsertId() error = %v", err)
	}
	return stageID
}

func seedTaskSearchTask(
	t *testing.T,
	db *sql.DB,
	userID string,
	stageID int64,
	title string,
	description string,
	deadline *time.Time,
	priority *string,
	assignedTo *string,
) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO tasks
			(user_id, stage_id, title, description, position, deadline, priority, assigned_to)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		stageID,
		title,
		description,
		0,
		nullableTimelineTime(deadline),
		nullableTimelineString(priority),
		nullableTimelineString(assignedTo),
	)
	if err != nil {
		t.Fatalf("insert task error = %v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("task LastInsertId() error = %v", err)
	}
	return taskID
}
