package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

// TaskLabelRepository handles database operations for task-label associations
type TaskLabelRepository struct {
	db *sql.DB
}

// NewTaskLabelRepository creates a new TaskLabelRepository
func NewTaskLabelRepository(db *sql.DB) *TaskLabelRepository {
	return &TaskLabelRepository{db: db}
}

// AssignLabelToTask assigns a label to a task
func (r *TaskLabelRepository) AssignLabelToTask(taskID, labelID int64) error {
	// Check if already assigned
	var exists int
	err := r.db.QueryRow(
		"SELECT 1 FROM task_labels WHERE task_id = ? AND label_id = ?",
		taskID, labelID,
	).Scan(&exists)

	if err == nil {
		return fmt.Errorf("label is already assigned to this task")
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing assignment: %v", err)
	}

	_, err = r.db.Exec(
		"INSERT INTO task_labels (task_id, label_id, created_at) VALUES (?, ?, ?)",
		taskID, labelID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to assign label: %v", err)
	}

	return nil
}

// RemoveLabelFromTask removes a label from a task
func (r *TaskLabelRepository) RemoveLabelFromTask(taskID, labelID int64) error {
	result, err := r.db.Exec(
		"DELETE FROM task_labels WHERE task_id = ? AND label_id = ?",
		taskID, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove label: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("label assignment not found")
	}

	return nil
}

// GetLabelsByTask retrieves all labels for a task
func (r *TaskLabelRepository) GetLabelsByTask(taskID int64) ([]models.Label, error) {
	rows, err := r.db.Query(`
		SELECT l.id, l.project_id, l.name, l.color, l.created_by, l.created_at
		FROM labels l
		INNER JOIN task_labels tl ON l.id = tl.label_id
		WHERE tl.task_id = ?
		ORDER BY l.name ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task labels: %v", err)
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var label models.Label
		err := rows.Scan(&label.ID, &label.ProjectID, &label.Name, &label.Color, &label.CreatedBy, &label.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %v", err)
		}
		labels = append(labels, label)
	}

	return labels, nil
}

// GetTaskIDsByLabel retrieves all task IDs that have a specific label
func (r *TaskLabelRepository) GetTaskIDsByLabel(labelID int64) ([]int64, error) {
	rows, err := r.db.Query(
		"SELECT task_id FROM task_labels WHERE label_id = ?",
		labelID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by label: %v", err)
	}
	defer rows.Close()

	var taskIDs []int64
	for rows.Next() {
		var taskID int64
		if err := rows.Scan(&taskID); err != nil {
			return nil, fmt.Errorf("failed to scan task ID: %v", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

// GetTaskLabel retrieves a task label association
func (r *TaskLabelRepository) GetTaskLabel(taskID, labelID int64) (*models.TaskLabel, error) {
	var tl models.TaskLabel
	err := r.db.QueryRow(
		"SELECT task_id, label_id FROM task_labels WHERE task_id = ? AND label_id = ?",
		taskID, labelID,
	).Scan(&tl.TaskID, &tl.LabelID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task label: %v", err)
	}

	return &tl, nil
}

// GetLabelsForTasks retrieves all labels for multiple tasks in a single query (bulk fetch)
func (r *TaskLabelRepository) GetLabelsForTasks(taskIDs []int64) (map[int64][]models.Label, error) {
	if len(taskIDs) == 0 {
		return make(map[int64][]models.Label), nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT l.id, l.project_id, l.name, l.color, l.created_by, l.created_at, tl.task_id
		FROM labels l
		INNER JOIN task_labels tl ON l.id = tl.label_id
		WHERE tl.task_id IN (%s)
		ORDER BY tl.task_id, l.name ASC
	`, placeholders)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for tasks: %v", err)
	}
	defer rows.Close()

	result := make(map[int64][]models.Label)
	for rows.Next() {
		var label models.Label
		var taskID int64
		err := rows.Scan(&label.ID, &label.ProjectID, &label.Name, &label.Color, &label.CreatedBy, &label.CreatedAt, &taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task label: %v", err)
		}
		result[taskID] = append(result[taskID], label)
	}

	return result, nil
}
