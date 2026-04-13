package testcases

import (
	"testing"
	"time"

	authmodels "backend/internal/auth/models"
	coretests "backend/internal/models"
)

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &authmodels.User{
		ID:           "user-123",
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "secret-hashed-password",
		Role:         authmodels.RoleUser,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	response := user.ToResponse()

	if response.ID != user.ID {
		t.Errorf("ToResponse() ID = %v, want %v", response.ID, user.ID)
	}
	if response.Name != user.Name {
		t.Errorf("ToResponse() Name = %v, want %v", response.Name, user.Name)
	}
	if response.Email != user.Email {
		t.Errorf("ToResponse() Email = %v, want %v", response.Email, user.Email)
	}
}

func TestUserRole_Constants(t *testing.T) {
	if authmodels.RoleUser != "user" {
		t.Errorf("RoleUser = %v, want user", authmodels.RoleUser)
	}
	if authmodels.RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %v, want admin", authmodels.RoleAdmin)
	}
}

func TestProject_Structure(t *testing.T) {
	now := time.Now()
	project := coretests.Project{
		ID:          1,
		OwnerID:     "user-123",
		Name:        "My Project",
		Description: "Project description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if project.ID != 1 {
		t.Errorf("Project.ID = %v, want 1", project.ID)
	}
	if project.Name != "My Project" {
		t.Errorf("Project.Name = %v, want My Project", project.Name)
	}
	if project.Description != "Project description" {
		t.Errorf("Project.Description = %v, want Project description", project.Description)
	}
}

func TestStage_Structure(t *testing.T) {
	now := time.Now()
	stage := coretests.Stage{
		ID:        1,
		UserID:    "user-123",
		ProjectID: 1,
		Name:      "To Do",
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if stage.Name != "To Do" {
		t.Errorf("Stage.Name = %v, want To Do", stage.Name)
	}
	if stage.Position != 0 {
		t.Errorf("Stage.Position = %v, want 0", stage.Position)
	}
}

func TestTask_Structure(t *testing.T) {
	now := time.Now()
	task := coretests.Task{
		ID:          1,
		UserID:      "user-123",
		StageID:     1,
		Title:       "Implement feature",
		Description: "Add new functionality",
		Position:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if task.Title != "Implement feature" {
		t.Errorf("Task.Title = %v, want Implement feature", task.Title)
	}
	if task.Description != "Add new functionality" {
		t.Errorf("Task.Description = %v, want Add new functionality", task.Description)
	}
}

func TestMessage_Structure(t *testing.T) {
	now := time.Now()
	message := coretests.Message{
		ID:         1,
		UserID:     "user-123",
		ProjectID:  1,
		SenderName: "John",
		Content:    "Hello, world!",
		CreatedAt:  now,
	}

	if message.Content != "Hello, world!" {
		t.Errorf("Message.Content = %v, want Hello, world!", message.Content)
	}
	if message.SenderName != "John" {
		t.Errorf("Message.SenderName = %v, want John", message.SenderName)
	}
}

func TestZeroValues(t *testing.T) {
	t.Run("zero Project", func(t *testing.T) {
		var project coretests.Project
		if project.ID != 0 {
			t.Errorf("zero Project.ID = %v, want 0", project.ID)
		}
		if project.Name != "" {
			t.Errorf("zero Project.Name = %v, want empty string", project.Name)
		}
	})

	t.Run("zero Stage", func(t *testing.T) {
		var stage coretests.Stage
		if stage.ID != 0 {
			t.Errorf("zero Stage.ID = %v, want 0", stage.ID)
		}
		if stage.Position != 0 {
			t.Errorf("zero Stage.Position = %v, want 0", stage.Position)
		}
	})

	t.Run("zero Task", func(t *testing.T) {
		var task coretests.Task
		if task.ID != 0 {
			t.Errorf("zero Task.ID = %v, want 0", task.ID)
		}
		if task.Title != "" {
			t.Errorf("zero Task.Title = %v, want empty string", task.Title)
		}
	})

	t.Run("zero Message", func(t *testing.T) {
		var message coretests.Message
		if message.ID != 0 {
			t.Errorf("zero Message.ID = %v, want 0", message.ID)
		}
		if message.Content != "" {
			t.Errorf("zero Message.Content = %v, want empty string", message.Content)
		}
	})
}
