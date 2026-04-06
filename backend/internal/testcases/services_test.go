package testcases

import (
	"testing"

	"backend/internal/services"
)

func TestNewProjectService(t *testing.T) {
	svc := services.NewProjectService(nil)
	if svc == nil {
		t.Error("NewProjectService() should not return nil")
	}
}

func TestNewStageService(t *testing.T) {
	svc := services.NewStageService(nil)
	if svc == nil {
		t.Error("NewStageService() should not return nil")
	}
}

func TestNewTaskService(t *testing.T) {
	svc := services.NewTaskService(nil)
	if svc == nil {
		t.Error("NewTaskService() should not return nil")
	}
}

func TestNewMessageService(t *testing.T) {
	svc := services.NewMessageService(nil)
	if svc == nil {
		t.Error("NewMessageService() should not return nil")
	}
}
