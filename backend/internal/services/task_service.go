package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"
)

type TaskService struct {
	db          *sql.DB
	activitySvc *ActivityService
}

var ErrInvalidTaskPriority = errors.New("invalid task priority")

func NewTaskService(db *sql.DB, activitySvc *ActivityService) *TaskService {
	return &TaskService{db: db, activitySvc: activitySvc}
}

var allowedTaskPriorities = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
	"urgent": {},
}

type TaskAttributes struct {
	StartDate  *time.Time
	Deadline   *time.Time
	Priority   *string
	AssignedTo *string
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
func (s *TaskService) CreateTask(userID string, stageID int64, title, description string, position int, attrs TaskAttributes) (*models.Task, error) {
	// Verify stage belongs to user's project
	projectID, err := s.verifyStageOwnership(userID, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify stage ownership: %v", err)
	}
	if projectID == 0 {
		return nil, fmt.Errorf("stage not found or access denied")
	}

	attrs, err = normalizeTaskAttributes(attrs)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(
		"INSERT INTO tasks (user_id, stage_id, title, description, position, start_date, deadline, priority, assigned_to) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		userID, stageID, title, description, position, nullableTime(attrs.StartDate), nullableTime(attrs.Deadline), nullableString(attrs.Priority), nullableString(attrs.AssignedTo),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	// Log activity (ignore errors, non-critical)
	if s.activitySvc != nil {
		s.activitySvc.LogTaskCreated(projectID, userID, "", id, title)
	}

	return &models.Task{
		ID:          id,
		UserID:      userID,
		StageID:     stageID,
		Title:       title,
		Description: description,
		Position:    position,
		StartDate:   attrs.StartDate,
		Deadline:    attrs.Deadline,
		Priority:    attrs.Priority,
		AssignedTo:  attrs.AssignedTo,
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
		`SELECT
			tasks.id,
			tasks.user_id,
			tasks.stage_id,
			tasks.title,
			tasks.description,
			tasks.position,
			tasks.start_date,
			tasks.deadline,
			tasks.priority,
			tasks.assigned_to,
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE subtasks.task_id = tasks.id), 0) AS subtask_count,
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE subtasks.task_id = tasks.id AND subtasks.is_completed = 1), 0) AS completed_count,
			tasks.created_at,
			tasks.updated_at
		FROM tasks
		WHERE tasks.stage_id = ? AND tasks.user_id = ?
		ORDER BY tasks.position`,
		stageID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %v", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID retrieves a task by ID (validates ownership)
func (s *TaskService) GetTaskByID(userID string, id int64) (*models.Task, error) {
	row := s.db.QueryRow(
		`SELECT
			tasks.id,
			tasks.user_id,
			tasks.stage_id,
			tasks.title,
			tasks.description,
			tasks.position,
			tasks.start_date,
			tasks.deadline,
			tasks.priority,
			tasks.assigned_to,
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE subtasks.task_id = tasks.id), 0) AS subtask_count,
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE subtasks.task_id = tasks.id AND subtasks.is_completed = 1), 0) AS completed_count,
			tasks.created_at,
			tasks.updated_at
		FROM tasks
		WHERE tasks.id = ? AND tasks.user_id = ?`,
		id, userID,
	)
	task, err := scanTask(row)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	return &task, nil
}

// UpdateTask updates a task (validates ownership)
func (s *TaskService) UpdateTask(userID string, id int64, title, description string, position int, attrs TaskAttributes) (*models.Task, error) {
	attrs, err := normalizeTaskAttributes(attrs)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(
		"UPDATE tasks SET title = ?, description = ?, position = ?, start_date = ?, deadline = ?, priority = ?, assigned_to = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		title, description, position, nullableTime(attrs.StartDate), nullableTime(attrs.Deadline), nullableString(attrs.Priority), nullableString(attrs.AssignedTo), id, userID,
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

	// Log activity
	if s.activitySvc != nil {
		s.activitySvc.LogTaskMoved(projectID, userID, "", id, "", "", "")
	}

	return s.GetTaskByID(userID, id)
}

// DeleteTask deletes a task (validates ownership)
func (s *TaskService) DeleteTask(userID string, id int64) error {
	// Get projectID for logging before deleting
	var projectID int64
	s.db.QueryRow(`
		SELECT stages.project_id 
		FROM tasks 
		JOIN stages ON tasks.stage_id = stages.id 
		WHERE tasks.id = ? AND tasks.user_id = ?`, id, userID).Scan(&projectID)

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

	// Log activity
	if s.activitySvc != nil && projectID > 0 {
		s.activitySvc.LogTaskDeleted(projectID, userID, "", "")
	}

	return nil
}

type taskScanner interface {
	Scan(dest ...interface{}) error
}

func scanTask(scanner taskScanner) (models.Task, error) {
	var task models.Task
	var startDate sql.NullTime
	var deadline sql.NullTime
	var priority sql.NullString
	var assignedTo sql.NullString

	err := scanner.Scan(
		&task.ID,
		&task.UserID,
		&task.StageID,
		&task.Title,
		&task.Description,
		&task.Position,
		&startDate,
		&deadline,
		&priority,
		&assignedTo,
		&task.SubtaskCount,
		&task.CompletedCount,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}

	task.StartDate = nullableTimePtr(startDate)
	task.Deadline = nullableTimePtr(deadline)
	task.Priority = nullableStringPtr(priority)
	task.AssignedTo = nullableStringPtr(assignedTo)

	return task, nil
}

func normalizeTaskAttributes(attrs TaskAttributes) (TaskAttributes, error) {
	if attrs.Priority != nil {
		normalized := strings.ToLower(strings.TrimSpace(*attrs.Priority))
		if normalized == "" {
			attrs.Priority = nil
		} else {
			if _, ok := allowedTaskPriorities[normalized]; !ok {
				return TaskAttributes{}, ErrInvalidTaskPriority
			}
			attrs.Priority = &normalized
		}
	}

	if attrs.AssignedTo != nil {
		normalized := strings.TrimSpace(*attrs.AssignedTo)
		if normalized == "" {
			attrs.AssignedTo = nil
		} else {
			attrs.AssignedTo = &normalized
		}
	}

	return attrs, nil
}

func nullableTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}
