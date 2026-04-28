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

func TestTaskService_GetProjectTimelineFiltersDatedTasks(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, todoStageID, doneStageID := seedTimelineProject(t, db, "user-1")
	deadline := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	priority := "high"
	assignedTo := "user-2"

	deadlineTaskID := seedTimelineTask(t, db, "user-1", todoStageID, "Deadline task", nil, &deadline, &priority, &assignedTo)
	startOnlyTaskID := seedTimelineTask(t, db, "user-1", doneStageID, "Start-only task", &startDate, nil, nil, nil)
	seedTimelineTask(t, db, "user-1", todoStageID, "Undated task", nil, nil, nil, nil)

	service := services.NewTaskService(db, nil)
	timeline, err := service.GetProjectTimeline("user-1", projectID)
	if err != nil {
		t.Fatalf("GetProjectTimeline() error = %v", err)
	}

	if len(timeline) != 2 {
		t.Fatalf("GetProjectTimeline() len = %d, want 2: %+v", len(timeline), timeline)
	}

	if timeline[0].TaskID != deadlineTaskID {
		t.Fatalf("first task ID = %d, want deadline task %d", timeline[0].TaskID, deadlineTaskID)
	}
	if timeline[0].StageName != "To Do" {
		t.Fatalf("first stage name = %q, want To Do", timeline[0].StageName)
	}
	if timeline[0].Deadline == nil || !timeline[0].Deadline.Equal(deadline) {
		t.Fatalf("first deadline = %v, want %v", timeline[0].Deadline, deadline)
	}
	if timeline[0].Priority == nil || *timeline[0].Priority != "high" {
		t.Fatalf("first priority = %v, want high", timeline[0].Priority)
	}
	if timeline[0].AssignedTo == nil || *timeline[0].AssignedTo != "user-2" {
		t.Fatalf("first assigned_to = %v, want user-2", timeline[0].AssignedTo)
	}

	if timeline[1].TaskID != startOnlyTaskID {
		t.Fatalf("second task ID = %d, want start-only task %d", timeline[1].TaskID, startOnlyTaskID)
	}
	if timeline[1].StageName != "Done" {
		t.Fatalf("second stage name = %q, want Done", timeline[1].StageName)
	}
	if timeline[1].StartDate == nil || !timeline[1].StartDate.Equal(startDate) {
		t.Fatalf("second start_date = %v, want %v", timeline[1].StartDate, startDate)
	}
	if timeline[1].Deadline != nil {
		t.Fatalf("second deadline = %v, want nil", timeline[1].Deadline)
	}
}

func TestTaskService_GetProjectTimelineReturnsEmptyArray(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTimelineProject(t, db, "user-1")
	seedTimelineTask(t, db, "user-1", stageID, "Undated task", nil, nil, nil, nil)

	service := services.NewTaskService(db, nil)
	timeline, err := service.GetProjectTimeline("user-1", projectID)
	if err != nil {
		t.Fatalf("GetProjectTimeline() error = %v", err)
	}
	if len(timeline) != 0 {
		t.Fatalf("GetProjectTimeline() len = %d, want 0", len(timeline))
	}
}

func TestTaskService_GetProjectTimelineOrdersDeadlinesBeforeStartOnlyTasks(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTimelineProject(t, db, "user-1")
	earlyDeadline := time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)
	lateDeadline := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	lateDeadlineID := seedTimelineTask(t, db, "user-1", stageID, "Late deadline", nil, &lateDeadline, nil, nil)
	startOnlyID := seedTimelineTask(t, db, "user-1", stageID, "Start only", &startDate, nil, nil, nil)
	earlyDeadlineID := seedTimelineTask(t, db, "user-1", stageID, "Early deadline", nil, &earlyDeadline, nil, nil)

	service := services.NewTaskService(db, nil)
	timeline, err := service.GetProjectTimeline("user-1", projectID)
	if err != nil {
		t.Fatalf("GetProjectTimeline() error = %v", err)
	}

	wantIDs := []int64{earlyDeadlineID, lateDeadlineID, startOnlyID}
	if len(timeline) != len(wantIDs) {
		t.Fatalf("GetProjectTimeline() len = %d, want %d: %+v", len(timeline), len(wantIDs), timeline)
	}
	for index, wantID := range wantIDs {
		if timeline[index].TaskID != wantID {
			t.Fatalf("timeline[%d].TaskID = %d, want %d; timeline=%+v", index, timeline[index].TaskID, wantID, timeline)
		}
	}
}

func TestTaskService_GetProjectTimelineRequiresProjectAccess(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTimelineProject(t, db, "user-1")
	deadline := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	seedTimelineTask(t, db, "user-1", stageID, "Private task", nil, &deadline, nil, nil)

	service := services.NewTaskService(db, nil)
	timeline, err := service.GetProjectTimeline("user-2", projectID)
	if err == nil {
		t.Fatal("GetProjectTimeline() error = nil, want access error")
	}
	if timeline != nil {
		t.Fatalf("GetProjectTimeline() = %+v, want nil", timeline)
	}
}

func TestTaskService_GetProjectTimelineAllowsProjectMember(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTimelineProject(t, db, "user-1")
	if _, err := db.Exec(
		"INSERT INTO project_members (project_id, user_id, role, invited_by) VALUES (?, ?, ?, ?)",
		projectID, "user-2", "member", "user-1",
	); err != nil {
		t.Fatalf("insert project member error = %v", err)
	}

	deadline := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	taskID := seedTimelineTask(t, db, "user-1", stageID, "Shared task", nil, &deadline, nil, nil)

	service := services.NewTaskService(db, nil)
	timeline, err := service.GetProjectTimeline("user-2", projectID)
	if err != nil {
		t.Fatalf("GetProjectTimeline() error = %v", err)
	}
	if len(timeline) != 1 || timeline[0].TaskID != taskID {
		t.Fatalf("GetProjectTimeline() = %+v, want shared task %d", timeline, taskID)
	}
}

func TestTaskController_GetProjectTimelineReturnsPlainArray(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, stageID, _ := seedTimelineProject(t, db, "user-1")
	deadline := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	taskID := seedTimelineTask(t, db, "user-1", stageID, "Deadline task", nil, &deadline, nil, nil)

	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/timeline", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetProjectTimeline(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetProjectTimeline() status = %d, want %d", w.Code, http.StatusOK)
	}

	var timeline []models.TimelineTaskResponse
	if err := json.NewDecoder(w.Body).Decode(&timeline); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if len(timeline) != 1 {
		t.Fatalf("response len = %d, want 1", len(timeline))
	}
	if timeline[0].TaskID != taskID {
		t.Fatalf("response task_id = %d, want %d", timeline[0].TaskID, taskID)
	}
}

func TestTaskController_GetProjectTimelineRejectsInvalidProjectID(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	req := createRequestWithUser(http.MethodGet, "/api/projects/not-a-number/timeline", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "not-a-number"})
	w := httptest.NewRecorder()

	controller.GetProjectTimeline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetProjectTimeline() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTaskController_GetProjectTimelineRejectsNegativeProjectID(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	req := createRequestWithUser(http.MethodGet, "/api/projects/-1/timeline", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "-1"})
	w := httptest.NewRecorder()

	controller.GetProjectTimeline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetProjectTimeline() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTaskController_GetProjectTimelineRequiresAuthentication(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/1/timeline", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	controller.GetProjectTimeline(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("GetProjectTimeline() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestTaskController_GetProjectTimelineMapsProjectErrors(t *testing.T) {
	db := newTimelineTestDB(t)
	defer db.Close()

	projectID, _, _ := seedTimelineProject(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)

	tests := []struct {
		name       string
		userID     string
		projectID  string
		wantStatus int
	}{
		{
			name:       "missing project",
			userID:     "user-1",
			projectID:  toString(projectID + 100),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no project access",
			userID:     "user-2",
			projectID:  toString(projectID),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithUser(http.MethodGet, "/api/projects/1/timeline", nil, tt.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.projectID})
			w := httptest.NewRecorder()

			controller.GetProjectTimeline(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("GetProjectTimeline() status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func newTimelineTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "timeline_*.db")
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

func seedTimelineProject(t *testing.T, db *sql.DB, userID string) (int64, int64, int64) {
	t.Helper()

	projectResult, err := db.Exec(
		"INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)",
		userID, "Project", "Description",
	)
	if err != nil {
		t.Fatalf("insert project error = %v", err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatalf("project LastInsertId() error = %v", err)
	}

	if _, err := db.Exec(
		"INSERT INTO project_members (project_id, user_id, role, invited_by) VALUES (?, ?, ?, ?)",
		projectID, userID, "owner", userID,
	); err != nil {
		t.Fatalf("insert project member error = %v", err)
	}

	todoStageID := seedTimelineStage(t, db, userID, projectID, "To Do", 0)
	doneStageID := seedTimelineStage(t, db, userID, projectID, "Done", 1)
	return projectID, todoStageID, doneStageID
}

func seedTimelineStage(t *testing.T, db *sql.DB, userID string, projectID int64, name string, position int) int64 {
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

func seedTimelineTask(
	t *testing.T,
	db *sql.DB,
	userID string,
	stageID int64,
	title string,
	startDate *time.Time,
	deadline *time.Time,
	priority *string,
	assignedTo *string,
) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO tasks
			(user_id, stage_id, title, description, position, start_date, deadline, priority, assigned_to)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		stageID,
		title,
		"Description",
		0,
		nullableTimelineTime(startDate),
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

func nullableTimelineTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimelineString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}
