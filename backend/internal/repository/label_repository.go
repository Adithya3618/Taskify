package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

// LabelRepository handles database operations for labels
type LabelRepository struct {
	db *sql.DB
}

// NewLabelRepository creates a new LabelRepository
func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

// CreateLabel creates a new label in a project
func (r *LabelRepository) CreateLabel(projectID int64, name, color, createdBy string) (*models.Label, error) {
	// Check for duplicate label name in project
	var existingID int64
	err := r.db.QueryRow(
		"SELECT id FROM labels WHERE project_id = ? AND name = ?",
		projectID, name,
	).Scan(&existingID)

	if err == nil {
		return nil, fmt.Errorf("label with name '%s' already exists in this project", name)
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing label: %v", err)
	}

	result, err := r.db.Exec(
		"INSERT INTO labels (project_id, name, color, created_by, created_at) VALUES (?, ?, ?, ?, ?)",
		projectID, name, color, createdBy, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return r.GetLabelByID(id)
}

// GetLabelsByProject retrieves all labels for a project
func (r *LabelRepository) GetLabelsByProject(projectID int64) ([]models.Label, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, name, color, created_by, created_at
		FROM labels
		WHERE project_id = ?
		ORDER BY created_at DESC, name ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %v", err)
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

// GetLabelByID retrieves a label by ID
func (r *LabelRepository) GetLabelByID(labelID int64) (*models.Label, error) {
	var label models.Label
	err := r.db.QueryRow(`
		SELECT id, project_id, name, color, created_by, created_at
		FROM labels
		WHERE id = ?
	`, labelID).Scan(&label.ID, &label.ProjectID, &label.Name, &label.Color, &label.CreatedBy, &label.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get label: %v", err)
	}

	return &label, nil
}

// DeleteLabel deletes a label (cascade will handle task_labels)
func (r *LabelRepository) DeleteLabel(labelID int64) error {
	result, err := r.db.Exec("DELETE FROM labels WHERE id = ?", labelID)
	if err != nil {
		return fmt.Errorf("failed to delete label: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("label not found")
	}

	return nil
}
