package testcases

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"backend/internal/controllers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func TestTaskService_CreateTask_BackwardCompatibleWithoutEnhancements(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)

	task, err := service.CreateTask("user-1", stageID, "Task", "desc", 0, services.TaskAttributes{})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if task.Deadline != nil {
		t.Errorf("Deadline = %v, want nil", task.Deadline)
	}
	if task.Priority != nil {
		t.Errorf("Priority = %v, want nil", task.Priority)
	}
	if task.AssignedTo != nil {
		t.Errorf("AssignedTo = %v, want nil", task.AssignedTo)
	}
	if task.StartDate != nil {
		t.Errorf("StartDate = %v, want nil", task.StartDate)
	}
}

func TestTaskService_CreateUpdateAndReadEnhancements(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)

	initialStartDate := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	initialDeadline := time.Date(2026, 4, 20, 15, 0, 0, 0, time.UTC)
	initialPriority := "high"
	initialAssignedTo := "user-2"

	created, err := service.CreateTask("user-1", stageID, "Enhanced Task", "desc", 1, services.TaskAttributes{
		StartDate:  &initialStartDate,
		Deadline:   &initialDeadline,
		Priority:   &initialPriority,
		AssignedTo: &initialAssignedTo,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	assertTaskEnhancements(t, created, &initialStartDate, &initialDeadline, "high", "user-2")

	fetched, err := service.GetTaskByID("user-1", created.ID)
	if err != nil {
		t.Fatalf("GetTaskByID() error = %v", err)
	}
	assertTaskEnhancements(t, fetched, &initialStartDate, &initialDeadline, "high", "user-2")

	listed, err := service.GetTasksByStage("user-1", stageID)
	if err != nil {
		t.Fatalf("GetTasksByStage() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("GetTasksByStage() len = %d, want 1", len(listed))
	}
	assertTaskEnhancements(t, &listed[0], &initialStartDate, &initialDeadline, "high", "user-2")

	updatedStartDate := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	updatedDeadline := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)
	updatedPriority := "urgent"
	updatedAssignedTo := "user-3"

	updated, err := service.UpdateTask("user-1", created.ID, "Updated Task", "updated desc", 2, services.TaskAttributes{
		StartDate:  &updatedStartDate,
		Deadline:   &updatedDeadline,
		Priority:   &updatedPriority,
		AssignedTo: &updatedAssignedTo,
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	assertTaskEnhancements(t, updated, &updatedStartDate, &updatedDeadline, "urgent", "user-3")

	cleared, err := service.UpdateTask("user-1", created.ID, "Updated Task", "updated desc", 2, services.TaskAttributes{})
	if err != nil {
		t.Fatalf("UpdateTask(clear) error = %v", err)
	}
	if cleared.StartDate != nil || cleared.Deadline != nil || cleared.Priority != nil || cleared.AssignedTo != nil {
		t.Errorf("expected cleared nullable fields, got %+v", cleared)
	}
}

func TestTaskService_InvalidPriorityRejected(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	invalidPriority := "critical"

	_, err := service.CreateTask("user-1", stageID, "Task", "desc", 0, services.TaskAttributes{
		Priority: &invalidPriority,
	})
	if err == nil {
		t.Fatal("CreateTask() error = nil, want invalid priority error")
	}
	if err != services.ErrInvalidTaskPriority {
		t.Fatalf("CreateTask() error = %v, want %v", err, services.ErrInvalidTaskPriority)
	}
}

func TestTaskService_InvalidDateRangeRejected(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)

	startDate := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

	_, err := service.CreateTask("user-1", stageID, "Task", "desc", 0, services.TaskAttributes{
		StartDate: &startDate,
		Deadline:  &deadline,
	})
	if err == nil {
		t.Fatal("CreateTask() error = nil, want invalid date range error")
	}
	if err != services.ErrInvalidDateRange {
		t.Fatalf("CreateTask() error = %v, want %v", err, services.ErrInvalidDateRange)
	}
}

func TestTaskController_InvalidPriorityReturnsBadRequest(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)

	t.Run("create rejects invalid priority", func(t *testing.T) {
		req := createRequestWithUser(http.MethodPost, "/api/projects/1/stages/1/tasks", map[string]interface{}{
			"title":       "Task",
			"description": "desc",
			"position":    0,
			"priority":    "critical",
		}, "user-1")
		req = mux.SetURLVars(req, map[string]string{"stageId": toString(stageID)})
		w := httptest.NewRecorder()

		controller.CreateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("CreateTask() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("update rejects invalid priority", func(t *testing.T) {
		taskID := seedTaskEnhancementTask(t, db, "user-1", stageID)
		req := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
			"title":       "Task",
			"description": "desc",
			"position":    0,
			"priority":    "critical",
		}, "user-1")
		req = mux.SetURLVars(req, map[string]string{"id": toString(taskID)})
		w := httptest.NewRecorder()

		controller.UpdateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("UpdateTask() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestTaskController_InvalidDateRangeReturnsBadRequest(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)

	t.Run("create rejects invalid date range", func(t *testing.T) {
		req := createRequestWithUser(http.MethodPost, "/api/projects/1/stages/1/tasks", map[string]interface{}{
			"title":       "Task",
			"description": "desc",
			"position":    0,
			"start_date":  "2026-05-10T09:00:00Z",
			"deadline":    "2026-05-01T09:00:00Z",
		}, "user-1")
		req = mux.SetURLVars(req, map[string]string{"stageId": toString(stageID)})
		w := httptest.NewRecorder()

		controller.CreateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("CreateTask() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("update rejects invalid date range", func(t *testing.T) {
		taskID := seedTaskEnhancementTask(t, db, "user-1", stageID)
		req := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
			"title":       "Task",
			"description": "desc",
			"position":    0,
			"start_date":  "2026-05-10T09:00:00Z",
			"deadline":    "2026-05-01T09:00:00Z",
		}, "user-1")
		req = mux.SetURLVars(req, map[string]string{"id": toString(taskID)})
		w := httptest.NewRecorder()

		controller.UpdateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("UpdateTask() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestTaskController_EnhancementRoundTripAndBackwardCompatibleUpdate(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)

	createStartDate := "2026-06-01T09:00:00Z"
	createDeadline := "2026-06-10T12:30:00Z"
	createReq := createRequestWithUser(http.MethodPost, "/api/projects/1/stages/1/tasks", map[string]interface{}{
		"title":       "Task with enhancements",
		"description": "desc",
		"position":    0,
		"start_date":  createStartDate,
		"deadline":    createDeadline,
		"priority":    " HIGH ",
		"assigned_to": "  user-2  ",
	}, "user-1")
	createReq = mux.SetURLVars(createReq, map[string]string{"stageId": toString(stageID)})
	createW := httptest.NewRecorder()

	controller.CreateTask(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("CreateTask() status = %d, want %d", createW.Code, http.StatusCreated)
	}

	var created models.Task
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response error = %v", err)
	}
	if created.Priority == nil || *created.Priority != "high" {
		t.Fatalf("created priority = %v, want high", created.Priority)
	}
	if created.AssignedTo == nil || *created.AssignedTo != "user-2" {
		t.Fatalf("created assigned_to = %v, want user-2", created.AssignedTo)
	}
	if created.StartDate == nil || created.StartDate.Format(time.RFC3339) != createStartDate {
		t.Fatalf("created start_date = %v, want %s", created.StartDate, createStartDate)
	}
	if created.Deadline == nil || created.Deadline.Format(time.RFC3339) != createDeadline {
		t.Fatalf("created deadline = %v, want %s", created.Deadline, createDeadline)
	}

	updateReq := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
		"title":       "Task after old-client update",
		"description": "updated desc",
		"position":    3,
	}, "user-1")
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": toString(created.ID)})
	updateW := httptest.NewRecorder()

	controller.UpdateTask(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("UpdateTask() status = %d, want %d", updateW.Code, http.StatusOK)
	}

	var updated models.Task
	if err := json.NewDecoder(updateW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update response error = %v", err)
	}
	if updated.Priority == nil || *updated.Priority != "high" {
		t.Fatalf("updated priority = %v, want preserved high", updated.Priority)
	}
	if updated.AssignedTo == nil || *updated.AssignedTo != "user-2" {
		t.Fatalf("updated assigned_to = %v, want preserved user-2", updated.AssignedTo)
	}
	if updated.StartDate == nil || updated.StartDate.Format(time.RFC3339) != createStartDate {
		t.Fatalf("updated start_date = %v, want preserved %s", updated.StartDate, createStartDate)
	}
	if updated.Deadline == nil || updated.Deadline.Format(time.RFC3339) != createDeadline {
		t.Fatalf("updated deadline = %v, want preserved %s", updated.Deadline, createDeadline)
	}

	clearReq := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
		"title":       "Task after clear",
		"description": "updated desc",
		"position":    4,
		"start_date":  nil,
		"deadline":    nil,
		"priority":    nil,
		"assigned_to": nil,
	}, "user-1")
	clearReq = mux.SetURLVars(clearReq, map[string]string{"id": toString(created.ID)})
	clearW := httptest.NewRecorder()

	controller.UpdateTask(clearW, clearReq)

	if clearW.Code != http.StatusOK {
		t.Fatalf("UpdateTask(clear) status = %d, want %d", clearW.Code, http.StatusOK)
	}

	getReq := createRequestWithUser(http.MethodGet, "/api/tasks/1", nil, "user-1")
	getReq = mux.SetURLVars(getReq, map[string]string{"id": toString(created.ID)})
	getW := httptest.NewRecorder()

	controller.GetTask(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GetTask() status = %d, want %d", getW.Code, http.StatusOK)
	}

	var fetched models.Task
	if err := json.NewDecoder(getW.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode get response error = %v", err)
	}
	if fetched.StartDate != nil || fetched.Deadline != nil || fetched.Priority != nil || fetched.AssignedTo != nil {
		t.Fatalf("expected cleared nullable fields, got %+v", fetched)
	}

	listReq := createRequestWithUser(http.MethodGet, "/api/projects/1/stages/1/tasks", nil, "user-1")
	listReq = mux.SetURLVars(listReq, map[string]string{"stageId": toString(stageID)})
	listW := httptest.NewRecorder()

	controller.GetTasksByStage(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("GetTasksByStage() status = %d, want %d", listW.Code, http.StatusOK)
	}

	var listed []models.Task
	if err := json.NewDecoder(listW.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list response error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("GetTasksByStage() len = %d, want 1", len(listed))
	}
	if listed[0].StartDate != nil || listed[0].Deadline != nil || listed[0].Priority != nil || listed[0].AssignedTo != nil {
		t.Fatalf("expected list response with cleared nullable fields, got %+v", listed[0])
	}
}

func TestTaskController_StartDateUpdateCanChangeAndClear(t *testing.T) {
	db := newTaskEnhancementTestDB(t)
	defer db.Close()

	stageID := seedTaskEnhancementStage(t, db, "user-1")
	service := services.NewTaskService(db, nil)
	controller := controllers.NewTaskController(service)
	taskID := seedTaskEnhancementTask(t, db, "user-1", stageID)
	firstStartDate := "2026-07-01T08:00:00Z"
	secondStartDate := "2026-07-03T08:00:00Z"

	setReq := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
		"title":       "Task with start date",
		"description": "desc",
		"position":    1,
		"start_date":  firstStartDate,
	}, "user-1")
	setReq = mux.SetURLVars(setReq, map[string]string{"id": toString(taskID)})
	setW := httptest.NewRecorder()

	controller.UpdateTask(setW, setReq)

	if setW.Code != http.StatusOK {
		t.Fatalf("UpdateTask(set start_date) status = %d, want %d", setW.Code, http.StatusOK)
	}
	var setTask models.Task
	if err := json.NewDecoder(setW.Body).Decode(&setTask); err != nil {
		t.Fatalf("decode set response error = %v", err)
	}
	if setTask.StartDate == nil || setTask.StartDate.Format(time.RFC3339) != firstStartDate {
		t.Fatalf("set start_date = %v, want %s", setTask.StartDate, firstStartDate)
	}

	changeReq := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
		"title":       "Task with changed start date",
		"description": "desc",
		"position":    2,
		"start_date":  secondStartDate,
	}, "user-1")
	changeReq = mux.SetURLVars(changeReq, map[string]string{"id": toString(taskID)})
	changeW := httptest.NewRecorder()

	controller.UpdateTask(changeW, changeReq)

	if changeW.Code != http.StatusOK {
		t.Fatalf("UpdateTask(change start_date) status = %d, want %d", changeW.Code, http.StatusOK)
	}
	var changedTask models.Task
	if err := json.NewDecoder(changeW.Body).Decode(&changedTask); err != nil {
		t.Fatalf("decode change response error = %v", err)
	}
	if changedTask.StartDate == nil || changedTask.StartDate.Format(time.RFC3339) != secondStartDate {
		t.Fatalf("changed start_date = %v, want %s", changedTask.StartDate, secondStartDate)
	}

	clearReq := createRequestWithUser(http.MethodPut, "/api/tasks/1", map[string]interface{}{
		"title":       "Task with cleared start date",
		"description": "desc",
		"position":    3,
		"start_date":  nil,
	}, "user-1")
	clearReq = mux.SetURLVars(clearReq, map[string]string{"id": toString(taskID)})
	clearW := httptest.NewRecorder()

	controller.UpdateTask(clearW, clearReq)

	if clearW.Code != http.StatusOK {
		t.Fatalf("UpdateTask(clear start_date) status = %d, want %d", clearW.Code, http.StatusOK)
	}
	var clearedTask models.Task
	if err := json.NewDecoder(clearW.Body).Decode(&clearedTask); err != nil {
		t.Fatalf("decode clear response error = %v", err)
	}
	if clearedTask.StartDate != nil {
		t.Fatalf("cleared start_date = %v, want nil", clearedTask.StartDate)
	}
}

func newTaskEnhancementTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "task_enhancements_*.db")
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
			start_date DATETIME,
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

func seedTaskEnhancementStage(t *testing.T, db *sql.DB, userID string) int64 {
	t.Helper()

	projectResult, err := db.Exec(
		"INSERT INTO projects (user_id, name, description) VALUES (?, ?, ?)",
		userID, "Project", "Description",
	)
	if err != nil {
		t.Fatalf("insert project error = %v", err)
	}

	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatalf("project LastInsertId() error = %v", err)
	}

	stageResult, err := db.Exec(
		"INSERT INTO stages (user_id, project_id, name, position) VALUES (?, ?, ?, ?)",
		userID, projectID, "To Do", 0,
	)
	if err != nil {
		t.Fatalf("insert stage error = %v", err)
	}

	stageID, err := stageResult.LastInsertId()
	if err != nil {
		t.Fatalf("stage LastInsertId() error = %v", err)
	}

	return stageID
}

func seedTaskEnhancementTask(t *testing.T, db *sql.DB, userID string, stageID int64) int64 {
	t.Helper()

	result, err := db.Exec(
		"INSERT INTO tasks (user_id, stage_id, title, description, position) VALUES (?, ?, ?, ?, ?)",
		userID, stageID, "Task", "desc", 0,
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

func assertTaskEnhancements(t *testing.T, task *models.Task, wantStartDate, wantDeadline *time.Time, wantPriority, wantAssignedTo string) {
	t.Helper()

	if task == nil {
		t.Fatal("task = nil")
	}
	if task.StartDate == nil || !task.StartDate.Equal(*wantStartDate) {
		t.Fatalf("StartDate = %v, want %v", task.StartDate, wantStartDate)
	}
	if task.Deadline == nil || !task.Deadline.Equal(*wantDeadline) {
		t.Fatalf("Deadline = %v, want %v", task.Deadline, wantDeadline)
	}
	if task.Priority == nil || *task.Priority != wantPriority {
		t.Fatalf("Priority = %v, want %q", task.Priority, wantPriority)
	}
	if task.AssignedTo == nil || *task.AssignedTo != wantAssignedTo {
		t.Fatalf("AssignedTo = %v, want %q", task.AssignedTo, wantAssignedTo)
	}
}

func toString(value int64) string {
	return strconv.FormatInt(value, 10)
}
