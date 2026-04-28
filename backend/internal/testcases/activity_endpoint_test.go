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

type activityFeedResponse struct {
	Logs  []activityFeedItem `json:"logs"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
}

type activityFeedItem struct {
	ID          int64  `json:"id"`
	UserName    string `json:"user_name"`
	Action      string `json:"action"`
	EntityType  string `json:"entity_type"`
	EntityTitle string `json:"entity_title"`
	CreatedAt   string `json:"created_at"`
}

func TestActivityController_GetActivityReturnsPaginatedFeed(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskCreated, models.EntityTask, 1, "Oldest task", time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskMoved, models.EntityTask, 2, "Middle task", time.Date(2026, 4, 21, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-2", "Member User", models.ActivityCommentAdded, models.EntityComment, 3, "Newest comment", time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID+1, "user-9", "Other User", models.ActivityMemberAdded, models.EntityMember, 4, "Other project", time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC))

	controller := newActivityEndpointController(db)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/activity?page=1&limit=2", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}

	if response.Total != 3 {
		t.Fatalf("response.Total = %d, want 3", response.Total)
	}
	if response.Page != 1 {
		t.Fatalf("response.Page = %d, want 1", response.Page)
	}
	if len(response.Logs) != 2 {
		t.Fatalf("len(response.Logs) = %d, want 2: %+v", len(response.Logs), response.Logs)
	}
	if response.Logs[0].Action != string(models.ActivityCommentAdded) {
		t.Fatalf("first action = %q, want newest comment action", response.Logs[0].Action)
	}
	if response.Logs[0].EntityTitle != "Newest comment" {
		t.Fatalf("first entity_title = %q, want Newest comment", response.Logs[0].EntityTitle)
	}
	if response.Logs[1].Action != string(models.ActivityTaskMoved) {
		t.Fatalf("second action = %q, want task moved action", response.Logs[1].Action)
	}
}

func TestActivityController_GetActivityReturnsSecondPage(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskCreated, models.EntityTask, 1, "First", time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskMoved, models.EntityTask, 2, "Second", time.Date(2026, 4, 21, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskUpdated, models.EntityTask, 3, "Third", time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC))

	controller := newActivityEndpointController(db)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/activity?page=2&limit=2", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if response.Total != 3 {
		t.Fatalf("response.Total = %d, want 3", response.Total)
	}
	if response.Page != 2 {
		t.Fatalf("response.Page = %d, want 2", response.Page)
	}
	if len(response.Logs) != 1 {
		t.Fatalf("len(response.Logs) = %d, want 1", len(response.Logs))
	}
	if response.Logs[0].EntityTitle != "First" {
		t.Fatalf("page 2 entity_title = %q, want First", response.Logs[0].EntityTitle)
	}
}

func TestActivityController_GetActivityNormalizesPagination(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskCreated, models.EntityTask, 1, "First", time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskMoved, models.EntityTask, 2, "Second", time.Date(2026, 4, 21, 9, 0, 0, 0, time.UTC))
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskUpdated, models.EntityTask, 3, "Third", time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC))

	controller := newActivityEndpointController(db)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/activity?page=0&limit=0", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if response.Page != 1 {
		t.Fatalf("response.Page = %d, want 1", response.Page)
	}
	if response.Total != 3 {
		t.Fatalf("response.Total = %d, want 3", response.Total)
	}
	if len(response.Logs) != 3 {
		t.Fatalf("len(response.Logs) = %d, want 3", len(response.Logs))
	}
	if response.Logs[0].EntityTitle != "Third" {
		t.Fatalf("first entity_title = %q, want Third", response.Logs[0].EntityTitle)
	}
}

func TestActivityController_GetActivityCapsLargeLimit(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	for i := 0; i < 105; i++ {
		seedActivityEndpointLog(
			t,
			db,
			projectID,
			"user-1",
			"Owner User",
			models.ActivityTaskUpdated,
			models.EntityTask,
			int64(i+1),
			"Activity item",
			time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC).Add(time.Duration(i)*time.Minute),
		)
	}

	controller := newActivityEndpointController(db)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/activity?page=1&limit=200", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if response.Total != 105 {
		t.Fatalf("response.Total = %d, want 105", response.Total)
	}
	if len(response.Logs) != 100 {
		t.Fatalf("len(response.Logs) = %d, want 100", len(response.Logs))
	}
}

func TestActivityController_GetActivityFiltersByUserAndDateRange(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	start := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	middle := time.Date(2026, 4, 21, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC)
	seedActivityEndpointLog(t, db, projectID, "user-1", "Owner User", models.ActivityTaskCreated, models.EntityTask, 1, "Outside earlier", start)
	seedActivityEndpointLog(t, db, projectID, "user-2", "Member User", models.ActivityTaskMoved, models.EntityTask, 2, "Matching member activity", middle)
	seedActivityEndpointLog(t, db, projectID, "user-2", "Member User", models.ActivityTaskUpdated, models.EntityTask, 3, "Outside later", end)
	seedActivityEndpointLog(t, db, projectID, "user-3", "Other User", models.ActivityCommentAdded, models.EntityComment, 4, "Wrong user", middle)

	controller := newActivityEndpointController(db)
	req := createRequestWithUser(
		http.MethodGet,
		"/api/projects/1/activity?user_id=user-2&from=2026-04-21T00:00:00Z&to=2026-04-21T23:59:59Z",
		nil,
		"user-1",
	)
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if response.Total != 1 {
		t.Fatalf("response.Total = %d, want 1", response.Total)
	}
	if len(response.Logs) != 1 {
		t.Fatalf("len(response.Logs) = %d, want 1: %+v", len(response.Logs), response.Logs)
	}
	if response.Logs[0].EntityTitle != "Matching member activity" {
		t.Fatalf("entity_title = %q, want Matching member activity", response.Logs[0].EntityTitle)
	}
}

func TestActivityController_GetActivityReturnsEmptyFeed(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	controller := newActivityEndpointController(db)
	req := createRequestWithUser(http.MethodGet, "/api/projects/1/activity?page=1&limit=20", nil, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.GetActivity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var response activityFeedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if len(response.Logs) != 0 {
		t.Fatalf("len(response.Logs) = %d, want 0", len(response.Logs))
	}
	if response.Total != 0 {
		t.Fatalf("response.Total = %d, want 0", response.Total)
	}
}

func TestActivityController_GetActivityValidation(t *testing.T) {
	db := newActivityEndpointTestDB(t)
	defer db.Close()

	projectID := seedActivityEndpointProject(t, db, "user-1")
	controller := newActivityEndpointController(db)

	tests := []struct {
		name       string
		url        string
		userID     string
		projectID  string
		wantStatus int
	}{
		{
			name:       "unauthenticated",
			url:        "/api/projects/1/activity?page=1&limit=20",
			userID:     "",
			projectID:  toString(projectID),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid project ID",
			url:        "/api/projects/abc/activity?page=1&limit=20",
			userID:     "user-1",
			projectID:  "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing project",
			url:        "/api/projects/99999/activity?page=1&limit=20",
			userID:     "user-1",
			projectID:  "99999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no project access",
			url:        "/api/projects/1/activity?page=1&limit=20",
			userID:     "user-2",
			projectID:  toString(projectID),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid page parameter",
			url:        "/api/projects/1/activity?page=abc&limit=20",
			userID:     "user-1",
			projectID:  toString(projectID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid limit parameter",
			url:        "/api/projects/1/activity?page=1&limit=xyz",
			userID:     "user-1",
			projectID:  toString(projectID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid from parameter",
			url:        "/api/projects/1/activity?from=not-a-date",
			userID:     "user-1",
			projectID:  toString(projectID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid to parameter",
			url:        "/api/projects/1/activity?to=also-not-a-date",
			userID:     "user-1",
			projectID:  toString(projectID),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithUser(http.MethodGet, tt.url, nil, tt.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.projectID})
			w := httptest.NewRecorder()

			controller.GetActivity(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("GetActivity() status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func newActivityEndpointTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "activity_endpoint_*.db")
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
		`CREATE TABLE activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			user_name TEXT,
			action TEXT NOT NULL,
			entity_type TEXT,
			entity_id INTEGER,
			description TEXT NOT NULL,
			details TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

func newActivityEndpointController(db *sql.DB) *controllers.ActivityController {
	memberService := services.NewProjectMemberService(db)
	activityService := services.NewActivityService(db, memberService)
	return controllers.NewActivityController(activityService)
}

func seedActivityEndpointProject(t *testing.T, db *sql.DB, ownerID string) int64 {
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
	return projectID
}

func seedActivityEndpointLog(
	t *testing.T,
	db *sql.DB,
	projectID int64,
	userID string,
	userName string,
	action models.ActivityAction,
	entityType models.EntityType,
	entityID int64,
	description string,
	createdAt time.Time,
) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO activity_logs
			(project_id, user_id, user_name, action, entity_type, entity_id, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		projectID,
		userID,
		userName,
		string(action),
		string(entityType),
		entityID,
		description,
		createdAt,
	)
	if err != nil {
		t.Fatalf("insert activity log error = %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("activity LastInsertId() error = %v", err)
	}
	return id
}
