package testcases

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/controllers"
	"backend/internal/helpers"
)

func createRequestWithUser(method, url string, body interface{}, userID string) *http.Request {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if userID != "" {
		ctx := context.WithValue(req.Context(), "user_id", userID)
		return req.WithContext(ctx)
	}
	return req
}

func TestProjectController_CreateProject_Unauthorized(t *testing.T) {
	ctrl := controllers.NewProjectController(nil)

	t.Run("returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := createRequestWithUser("POST", "/api/projects", map[string]string{
			"name":        "Test Project",
			"description": "Test Description",
		}, "")

		ctrl.CreateProject(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("CreateProject() status = %v, want 401", w.Code)
		}
	})
}

func TestProjectController_CreateProject_InvalidJSON(t *testing.T) {
	ctrl := controllers.NewProjectController(nil)

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/projects", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), "user_id", "user-123")
		req = req.WithContext(ctx)

		ctrl.CreateProject(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("CreateProject() status = %v, want 400", w.Code)
		}
	})
}

func TestProjectController_GetAllProjects_Unauthorized(t *testing.T) {
	ctrl := controllers.NewProjectController(nil)

	t.Run("returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/projects", nil)

		ctrl.GetAllProjects(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetAllProjects() status = %v, want 401", w.Code)
		}
	})
}

func TestStageController_Unauthorized(t *testing.T) {
	stageCtrl := controllers.NewStageController(nil)

	t.Run("CreateStage returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/projects/1/stages", nil)

		stageCtrl.CreateStage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("CreateStage() status = %v, want 401", w.Code)
		}
	})

	t.Run("GetStagesByProject returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/projects/1/stages", nil)

		stageCtrl.GetStagesByProject(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetStagesByProject() status = %v, want 401", w.Code)
		}
	})

	t.Run("GetStage returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/stages/1", nil)

		stageCtrl.GetStage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetStage() status = %v, want 401", w.Code)
		}
	})
}

func TestTaskController_Unauthorized(t *testing.T) {
	taskCtrl := controllers.NewTaskController(nil)

	t.Run("CreateTask returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/stages/1/tasks", nil)

		taskCtrl.CreateTask(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("CreateTask() status = %v, want 401", w.Code)
		}
	})

	t.Run("GetTasksByStage returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/stages/1/tasks", nil)

		taskCtrl.GetTasksByStage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetTasksByStage() status = %v, want 401", w.Code)
		}
	})

	t.Run("GetTask returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/tasks/1", nil)

		taskCtrl.GetTask(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetTask() status = %v, want 401", w.Code)
		}
	})
}

func TestMessageController_Unauthorized(t *testing.T) {
	msgCtrl := controllers.NewMessageController(nil)

	t.Run("CreateMessage returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/projects/1/messages", nil)

		msgCtrl.CreateMessage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("CreateMessage() status = %v, want 401", w.Code)
		}
	})

	t.Run("GetMessagesByProject returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/projects/1/messages", nil)

		msgCtrl.GetMessagesByProject(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetMessagesByProject() status = %v, want 401", w.Code)
		}
	})

	t.Run("DeleteMessage returns 401 without user context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/messages/1", nil)

		msgCtrl.DeleteMessage(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("DeleteMessage() status = %v, want 401", w.Code)
		}
	})
}

func TestNewProjectController(t *testing.T) {
	ctrl := controllers.NewProjectController(nil)
	if ctrl == nil {
		t.Error("NewProjectController() should not return nil")
	}
}

func TestNewStageController(t *testing.T) {
	ctrl := controllers.NewStageController(nil)
	if ctrl == nil {
		t.Error("NewStageController() should not return nil")
	}
}

func TestNewTaskController(t *testing.T) {
	ctrl := controllers.NewTaskController(nil)
	if ctrl == nil {
		t.Error("NewTaskController() should not return nil")
	}
}

func TestNewMessageController(t *testing.T) {
	ctrl := controllers.NewMessageController(nil)
	if ctrl == nil {
		t.Error("NewMessageController() should not return nil")
	}
}

func TestGetUserID_Helper(t *testing.T) {
	t.Run("returns user ID from request", func(t *testing.T) {
		req := createRequestWithUser("GET", "/test", nil, "user-123")
		result := helpers.GetUserID(req)
		if result != "user-123" {
			t.Errorf("GetUserID() = %v, want user-123", result)
		}
	})

	t.Run("returns empty string without context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		result := helpers.GetUserID(req)
		if result != "" {
			t.Errorf("GetUserID() = %v, want empty string", result)
		}
	})
}
