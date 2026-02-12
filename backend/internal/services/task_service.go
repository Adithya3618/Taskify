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

// CreateTask creates a new task in a stage
func (s *TaskService) CreateTask(stageID int64, title, description string, position int) (*models.Task, error) {
	result, err := s.db.Exec(
		"INSERT INTO tasks (stage_id, title, description, position) VALUES (?, ?, ?, ?)",
		stageID, title, description, position,
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
		StageID:     stageID,
		Title:       title,
		Description: description,
		Position:    position,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetTasksByStage retrieves all tasks for a stage
func (s *TaskService) GetTasksByStage(stageID int64) ([]models.Task, error) {
	rows, err := s.db.Query(
		"SELECT id, stage_id, title, description, position, created_at, updated_at FROM tasks WHERE stage_id = ? ORDER BY position",
		stageID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %v", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.StageID, &task.Title, &task.Description, &task.Position, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID retrieves a task by ID
func (s *TaskService) GetTaskByID(id int64) (*models.Task, error) {
	var task models.Task
	err := s.db.QueryRow(
		"SELECT id, stage_id, title, description, position, created_at, updated_at FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.StageID, &task.Title, &task.Description, &task.Position, &task.CreatedAt, &task.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	return &task, nil
}

// UpdateTask updates a task
func (s *TaskService) UpdateTask(id int64, title, description string, position int) (*models.Task, error) {
	_, err := s.db.Exec(
		"UPDATE tasks SET title = ?, description = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		title, description, position, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %v", err)
	}

	return s.GetTaskByID(id)
}

// MoveTask moves a task to a different stage
func (s *TaskService) MoveTask(id int64, newStageID int64, newPosition int) (*models.Task, error) {
	_, err := s.db.Exec(
		"UPDATE tasks SET stage_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newStageID, newPosition, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to move task: %v", err)
	}

	return s.GetTaskByID(id)
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(id int64) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	return nil
}