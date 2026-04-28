package testcases

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/internal/controllers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func TestStageService_ReorderStagesUpdatesPositions(t *testing.T) {
	db := newStageReorderTestDB(t)
	defer db.Close()

	projectID, stageIDs := seedStageReorderProject(t, db, "user-1")
	service := services.NewStageService(db)

	reordered, err := service.ReorderStages("user-1", projectID, []int64{stageIDs[2], stageIDs[0], stageIDs[1]})
	if err != nil {
		t.Fatalf("ReorderStages() error = %v", err)
	}
	assertStageOrder(t, reordered, []int64{stageIDs[2], stageIDs[0], stageIDs[1]})

	persisted, err := service.GetStagesByProject("user-1", projectID)
	if err != nil {
		t.Fatalf("GetStagesByProject() error = %v", err)
	}
	assertStageOrder(t, persisted, []int64{stageIDs[2], stageIDs[0], stageIDs[1]})
}

func TestStageService_ReorderStagesRejectsInvalidRequests(t *testing.T) {
	db := newStageReorderTestDB(t)
	defer db.Close()

	projectID, stageIDs := seedStageReorderProject(t, db, "user-1")
	otherProjectID, otherStageIDs := seedStageReorderProject(t, db, "user-2")
	service := services.NewStageService(db)

	tests := []struct {
		name      string
		userID    string
		projectID int64
		stageIDs  []int64
		wantCode  string
	}{
		{
			name:      "empty stage list",
			userID:    "user-1",
			projectID: projectID,
			stageIDs:  []int64{},
			wantCode:  "INVALID_REQUEST",
		},
		{
			name:      "duplicate stage IDs",
			userID:    "user-1",
			projectID: projectID,
			stageIDs:  []int64{stageIDs[0], stageIDs[0], stageIDs[1]},
			wantCode:  "INVALID_REQUEST",
		},
		{
			name:      "stage belongs to another project",
			userID:    "user-1",
			projectID: projectID,
			stageIDs:  []int64{stageIDs[0], otherStageIDs[0], stageIDs[1]},
			wantCode:  "INVALID_REQUEST",
		},
		{
			name:      "project does not exist",
			userID:    "user-1",
			projectID: 99999,
			stageIDs:  []int64{stageIDs[0], stageIDs[1]},
			wantCode:  "PROJECT_NOT_FOUND",
		},
		{
			name:      "user has no project access",
			userID:    "user-3",
			projectID: otherProjectID,
			stageIDs:  []int64{otherStageIDs[0], otherStageIDs[1]},
			wantCode:  "ACCESS_DENIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ReorderStages(tt.userID, tt.projectID, tt.stageIDs)
			if err == nil {
				t.Fatal("ReorderStages() error = nil, want service error")
			}
			serviceErr, ok := services.IsServiceError(err)
			if !ok {
				t.Fatalf("ReorderStages() error = %T %v, want ServiceError", err, err)
			}
			if serviceErr.Code != tt.wantCode {
				t.Fatalf("ServiceError.Code = %q, want %q", serviceErr.Code, tt.wantCode)
			}
		})
	}
}

func TestStageService_ReorderStagesRollsBackInvalidOrder(t *testing.T) {
	db := newStageReorderTestDB(t)
	defer db.Close()

	projectID, stageIDs := seedStageReorderProject(t, db, "user-1")
	_, otherStageIDs := seedStageReorderProject(t, db, "user-2")
	service := services.NewStageService(db)

	_, err := service.ReorderStages("user-1", projectID, []int64{stageIDs[2], otherStageIDs[0], stageIDs[0]})
	if err == nil {
		t.Fatal("ReorderStages() error = nil, want invalid stage error")
	}

	persisted, err := service.GetStagesByProject("user-1", projectID)
	if err != nil {
		t.Fatalf("GetStagesByProject() error = %v", err)
	}
	assertStageOrder(t, persisted, []int64{stageIDs[0], stageIDs[1], stageIDs[2]})
}

func TestStageController_ReorderStagesReturnsUpdatedOrder(t *testing.T) {
	db := newStageReorderTestDB(t)
	defer db.Close()

	projectID, stageIDs := seedStageReorderProject(t, db, "user-1")
	service := services.NewStageService(db)
	controller := controllers.NewStageController(service)
	req := createRequestWithUser(http.MethodPut, "/api/projects/1/stages/reorder", map[string]interface{}{
		"stage_ids": []int64{stageIDs[2], stageIDs[1], stageIDs[0]},
	}, "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": toString(projectID)})
	w := httptest.NewRecorder()

	controller.ReorderStages(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ReorderStages() status = %d, want %d", w.Code, http.StatusOK)
	}

	var stages []models.Stage
	if err := json.NewDecoder(w.Body).Decode(&stages); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	assertStageOrder(t, stages, []int64{stageIDs[2], stageIDs[1], stageIDs[0]})
}

func TestStageController_ReorderStagesValidation(t *testing.T) {
	db := newStageReorderTestDB(t)
	defer db.Close()

	projectID, stageIDs := seedStageReorderProject(t, db, "user-1")
	_, otherStageIDs := seedStageReorderProject(t, db, "user-2")
	service := services.NewStageService(db)
	controller := controllers.NewStageController(service)

	tests := []struct {
		name       string
		userID     string
		projectID  string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "unauthenticated",
			userID:     "",
			projectID:  toString(projectID),
			body:       map[string]interface{}{"stage_ids": stageIDs},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid project ID",
			userID:     "user-1",
			projectID:  "abc",
			body:       map[string]interface{}{"stage_ids": stageIDs},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON body",
			userID:     "user-1",
			projectID:  toString(projectID),
			body:       rawJSON("not-json"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty stage list",
			userID:     "user-1",
			projectID:  toString(projectID),
			body:       map[string]interface{}{"stage_ids": []int64{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duplicate stage IDs",
			userID:     "user-1",
			projectID:  toString(projectID),
			body:       map[string]interface{}{"stage_ids": []int64{stageIDs[0], stageIDs[0]}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "stage from another project",
			userID:     "user-1",
			projectID:  toString(projectID),
			body:       map[string]interface{}{"stage_ids": []int64{stageIDs[0], otherStageIDs[0]}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown project",
			userID:     "user-1",
			projectID:  "99999",
			body:       map[string]interface{}{"stage_ids": stageIDs},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no project access",
			userID:     "user-3",
			projectID:  toString(projectID),
			body:       map[string]interface{}{"stage_ids": stageIDs},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := stageReorderRequest(tt.userID, tt.projectID, tt.body)
			w := httptest.NewRecorder()

			controller.ReorderStages(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("ReorderStages() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func newStageReorderTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "stage_reorder_*.db")
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

func seedStageReorderProject(t *testing.T, db *sql.DB, ownerID string) (int64, []int64) {
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

	stageIDs := []int64{
		seedStageReorderStage(t, db, ownerID, projectID, "Backlog", 0),
		seedStageReorderStage(t, db, ownerID, projectID, "Doing", 1),
		seedStageReorderStage(t, db, ownerID, projectID, "Done", 2),
	}
	return projectID, stageIDs
}

func seedStageReorderStage(t *testing.T, db *sql.DB, userID string, projectID int64, name string, position int) int64 {
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

type rawJSON string

func stageReorderRequest(userID, projectID string, body interface{}) *http.Request {
	var req *http.Request
	if raw, ok := body.(rawJSON); ok {
		req = httptest.NewRequest(http.MethodPut, "/api/projects/1/stages/reorder", strings.NewReader(string(raw)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = createRequestWithUser(http.MethodPut, "/api/projects/1/stages/reorder", body, userID)
	}
	if userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	}
	return mux.SetURLVars(req, map[string]string{"id": projectID})
}

func assertStageOrder(t *testing.T, stages []models.Stage, wantIDs []int64) {
	t.Helper()

	if len(stages) != len(wantIDs) {
		t.Fatalf("len(stages) = %d, want %d: %+v", len(stages), len(wantIDs), stages)
	}
	for index, wantID := range wantIDs {
		if stages[index].ID != wantID {
			t.Fatalf("stage[%d].ID = %d, want %d", index, stages[index].ID, wantID)
		}
		if stages[index].Position != index {
			t.Fatalf("stage[%d].Position = %d, want %d", index, stages[index].Position, index)
		}
	}
}
