package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend/internal/models"
)

var (
	ErrSubtaskNotFoundOrAccessDenied = errors.New("subtask not found or access denied")
	ErrSubtaskTitleRequired          = errors.New("subtask title is required")
	ErrInvalidSubtaskPosition        = errors.New("invalid subtask position")
)

type SubtaskService struct {
	db *sql.DB
}

type SubtaskPatch struct {
	Title       *string
	IsCompleted *bool
	Position    *int
}

func NewSubtaskService(db *sql.DB) *SubtaskService {
	return &SubtaskService{db: db}
}

func (s *SubtaskService) CreateSubtask(userID string, taskID int64, title string, position *int) (*models.Subtask, error) {
	normalizedTitle, err := normalizeSubtaskTitle(title)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	if _, err := s.verifyTaskOwnership(tx, userID, taskID); err != nil {
		return nil, err
	}

	count, err := s.countSubtasks(tx, taskID)
	if err != nil {
		return nil, err
	}

	targetPosition, err := normalizeInsertPosition(position, count)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		"UPDATE subtasks SET position = position + 1, updated_at = CURRENT_TIMESTAMP WHERE task_id = ? AND position >= ?",
		taskID, targetPosition,
	); err != nil {
		return nil, fmt.Errorf("failed to shift subtasks for insert: %v", err)
	}

	result, err := tx.Exec(
		"INSERT INTO subtasks (task_id, title, is_completed, position) VALUES (?, ?, ?, ?)",
		taskID, normalizedTitle, false, targetPosition,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subtask: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return s.GetSubtaskByID(userID, id)
}

func (s *SubtaskService) GetSubtasksByTask(userID string, taskID int64) ([]models.Subtask, error) {
	if _, err := s.verifyTaskOwnership(s.db, userID, taskID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		"SELECT id, task_id, title, is_completed, position, created_at, updated_at FROM subtasks WHERE task_id = ? ORDER BY position ASC, id ASC",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query subtasks: %v", err)
	}
	defer rows.Close()

	var subtasks []models.Subtask
	for rows.Next() {
		subtask, err := scanSubtask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subtask: %v", err)
		}
		subtasks = append(subtasks, subtask)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading subtasks: %v", err)
	}

	return subtasks, nil
}

func (s *SubtaskService) GetSubtaskByID(userID string, subtaskID int64) (*models.Subtask, error) {
	row := s.db.QueryRow(`
		SELECT subtasks.id, subtasks.task_id, subtasks.title, subtasks.is_completed, subtasks.position, subtasks.created_at, subtasks.updated_at
		FROM subtasks
		JOIN tasks ON subtasks.task_id = tasks.id
		JOIN stages ON tasks.stage_id = stages.id
		JOIN projects ON stages.project_id = projects.id
		WHERE subtasks.id = ? AND projects.owner_id = ?`,
		subtaskID, userID,
	)

	subtask, err := scanSubtask(row)
	if err == sql.ErrNoRows {
		return nil, ErrSubtaskNotFoundOrAccessDenied
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %v", err)
	}

	return &subtask, nil
}

func (s *SubtaskService) UpdateSubtask(userID string, subtaskID int64, patch SubtaskPatch) (*models.Subtask, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	current, err := s.getSubtaskByIDForUpdate(tx, userID, subtaskID)
	if err != nil {
		return nil, err
	}

	nextTitle := current.Title
	if patch.Title != nil {
		nextTitle, err = normalizeSubtaskTitle(*patch.Title)
		if err != nil {
			return nil, err
		}
	}

	nextCompleted := current.IsCompleted
	if patch.IsCompleted != nil {
		nextCompleted = *patch.IsCompleted
	}

	nextPosition := current.Position
	if patch.Position != nil {
		count, err := s.countSubtasks(tx, current.TaskID)
		if err != nil {
			return nil, err
		}
		nextPosition, err = normalizeMovePosition(*patch.Position, count)
		if err != nil {
			return nil, err
		}
		if nextPosition != current.Position {
			if err := s.reorderForMove(tx, current.TaskID, current.Position, nextPosition); err != nil {
				return nil, err
			}
		}
	}

	if _, err := tx.Exec(
		"UPDATE subtasks SET title = ?, is_completed = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		nextTitle, nextCompleted, nextPosition, subtaskID,
	); err != nil {
		return nil, fmt.Errorf("failed to update subtask: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return s.GetSubtaskByID(userID, subtaskID)
}

func (s *SubtaskService) DeleteSubtask(userID string, subtaskID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	current, err := s.getSubtaskByIDForUpdate(tx, userID, subtaskID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM subtasks WHERE id = ?", subtaskID); err != nil {
		return fmt.Errorf("failed to delete subtask: %v", err)
	}

	if _, err := tx.Exec(
		"UPDATE subtasks SET position = position - 1, updated_at = CURRENT_TIMESTAMP WHERE task_id = ? AND position > ?",
		current.TaskID, current.Position,
	); err != nil {
		return fmt.Errorf("failed to compact subtask positions: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

type subtaskScanner interface {
	Scan(dest ...interface{}) error
}

func scanSubtask(scanner subtaskScanner) (models.Subtask, error) {
	var subtask models.Subtask
	if err := scanner.Scan(
		&subtask.ID,
		&subtask.TaskID,
		&subtask.Title,
		&subtask.IsCompleted,
		&subtask.Position,
		&subtask.CreatedAt,
		&subtask.UpdatedAt,
	); err != nil {
		return models.Subtask{}, err
	}
	return subtask, nil
}

type queryable interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func (s *SubtaskService) verifyTaskOwnership(q queryable, userID string, taskID int64) (int64, error) {
	var taskIDFound int64
	err := q.QueryRow(`
		SELECT tasks.id
		FROM tasks
		JOIN stages ON tasks.stage_id = stages.id
		JOIN projects ON stages.project_id = projects.id
		WHERE tasks.id = ? AND projects.owner_id = ?`,
		taskID, userID,
	).Scan(&taskIDFound)
	if err == sql.ErrNoRows {
		return 0, ErrTaskNotFoundOrAccessDenied
	}
	if err != nil {
		return 0, fmt.Errorf("failed to verify task ownership: %v", err)
	}
	return taskIDFound, nil
}

func (s *SubtaskService) getSubtaskByIDForUpdate(tx *sql.Tx, userID string, subtaskID int64) (*models.Subtask, error) {
	row := tx.QueryRow(`
		SELECT subtasks.id, subtasks.task_id, subtasks.title, subtasks.is_completed, subtasks.position, subtasks.created_at, subtasks.updated_at
		FROM subtasks
		JOIN tasks ON subtasks.task_id = tasks.id
		JOIN stages ON tasks.stage_id = stages.id
		JOIN projects ON stages.project_id = projects.id
		WHERE subtasks.id = ? AND projects.owner_id = ?`,
		subtaskID, userID,
	)

	subtask, err := scanSubtask(row)
	if err == sql.ErrNoRows {
		return nil, ErrSubtaskNotFoundOrAccessDenied
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask for update: %v", err)
	}

	return &subtask, nil
}

func (s *SubtaskService) countSubtasks(tx *sql.Tx, taskID int64) (int, error) {
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM subtasks WHERE task_id = ?", taskID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count subtasks: %v", err)
	}
	return count, nil
}

func (s *SubtaskService) reorderForMove(tx *sql.Tx, taskID int64, oldPosition, newPosition int) error {
	if newPosition < oldPosition {
		if _, err := tx.Exec(
			"UPDATE subtasks SET position = position + 1, updated_at = CURRENT_TIMESTAMP WHERE task_id = ? AND position >= ? AND position < ?",
			taskID, newPosition, oldPosition,
		); err != nil {
			return fmt.Errorf("failed to reorder subtasks upward: %v", err)
		}
		return nil
	}

	if _, err := tx.Exec(
		"UPDATE subtasks SET position = position - 1, updated_at = CURRENT_TIMESTAMP WHERE task_id = ? AND position > ? AND position <= ?",
		taskID, oldPosition, newPosition,
	); err != nil {
		return fmt.Errorf("failed to reorder subtasks downward: %v", err)
	}
	return nil
}

func normalizeSubtaskTitle(title string) (string, error) {
	normalized := strings.TrimSpace(title)
	if normalized == "" {
		return "", ErrSubtaskTitleRequired
	}
	return normalized, nil
}

func normalizeInsertPosition(position *int, count int) (int, error) {
	if position == nil {
		return count, nil
	}
	if *position < 0 {
		return 0, ErrInvalidSubtaskPosition
	}
	if *position > count {
		return count, nil
	}
	return *position, nil
}

func normalizeMovePosition(position int, count int) (int, error) {
	if position < 0 {
		return 0, ErrInvalidSubtaskPosition
	}
	if count <= 0 {
		return 0, ErrInvalidSubtaskPosition
	}
	maxPosition := count - 1
	if position > maxPosition {
		return maxPosition, nil
	}
	return position, nil
}
