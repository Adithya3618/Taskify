package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

type TaskService struct {
	db *sql.DB
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{db: db}
}

// verifyStageOwnership checks if stage belongs to user's project
func (s *TaskService) verifyStageOwnership(userID string, stageID int64) (int64, error) {
	var projectID int64
	err := s.db.QueryRow(`
		SELECT stages.project_id 
		FROM stages 
		JOIN projects ON stages.project_id = projects.id 
		WHERE stages.id = ? AND projects.user_id = ?`, stageID, userID).Scan(&projectID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return projectID, nil
}

// CreateTask creates a new task in a stage (validates ownership)
func (s *TaskService) CreateTask(userID string, stageID int64, title, description string, position int) (*models.Task, error) {
	// Verify stage belongs to user's project
	projectID, err := s.verifyStageOwnership(userID, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify stage ownership: %v", err)
	}
	if projectID == 0 {
		return nil, fmt.Errorf("stage not found or access denied")
	}

	result, err := s.db.Exec(
		"INSERT INTO tasks (user_id, stage_id, title, description, position) VALUES (?, ?, ?, ?, ?)",
		userID, stageID, title, description, position,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return &models.Task{
		ID:          id,
		UserID:      userID,
		StageID:     stageID,
		Title:       title,
		Description: description,
		Position:    position,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetTasksByStage retrieves all tasks for a stage (validates ownership)
func (s *TaskService) GetTasksByStage(userID string, stageID int64) ([]models.Task, error) {
	// Verify stage belongs to user's project
	projectID, err := s.verifyStageOwnership(userID, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify stage ownership: %v", err)
	}
	if projectID == 0 {
		return nil, fmt.Errorf("stage not found or access denied")
	}

	rows, err := s.db.Query(
		"SELECT id, user_id, stage_id, title, description, position, created_at, updated_at FROM tasks WHERE stage_id = ? AND user_id = ? ORDER BY position",
		stageID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %v", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.StageID, &task.Title, &task.Description, &task.Position, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID retrieves a task by ID (validates ownership)
func (s *TaskService) GetTaskByID(userID string, id int64) (*models.Task, error) {
	var task models.Task
	err := s.db.QueryRow(
		"SELECT id, user_id, stage_id, title, description, position, created_at, updated_at FROM tasks WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&task.ID, &task.UserID, &task.StageID, &task.Title, &task.Description, &task.Position, &task.CreatedAt, &task.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	return &task, nil
}

// UpdateTask updates a task (validates ownership)
func (s *TaskService) UpdateTask(userID string, id int64, title, description string, position int) (*models.Task, error) {
	_, err := s.db.Exec(
		"UPDATE tasks SET title = ?, description = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		title, description, position, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %v", err)
	}

	return s.GetTaskByID(userID, id)
}

// MoveTask moves a task to a different stage (validates ownership)
func (s *TaskService) MoveTask(userID string, id int64, newStageID int64, newPosition int) (*models.Task, error) {
	// Verify new stage belongs to user's project
	projectID, err := s.verifyStageOwnership(userID, newStageID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify stage ownership: %v", err)
	}
	if projectID == 0 {
		return nil, fmt.Errorf("stage not found or access denied")
	}

	// Verify task belongs to user
	_, err = s.GetTaskByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	_, err = s.db.Exec(
		"UPDATE tasks SET stage_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		newStageID, newPosition, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to move task: %v", err)
	}

	return s.GetTaskByID(userID, id)
}

// DeleteTask deletes a task (validates ownership)
func (s *TaskService) DeleteTask(userID string, id int64) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found or access denied")
	}

	return nil
}
