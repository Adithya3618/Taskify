package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

type StageService struct {
	db *sql.DB
}

func NewStageService(db *sql.DB) *StageService {
	return &StageService{db: db}
}

// CreateStage creates a new stage for a project
func (s *StageService) CreateStage(projectID int64, name string, position int) (*models.Stage, error) {
	result, err := s.db.Exec(
		"INSERT INTO stages (project_id, name, position) VALUES (?, ?, ?)",
		projectID, name, position,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stage: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return &models.Stage{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		Position:  position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetStagesByProject retrieves all stages for a project
func (s *StageService) GetStagesByProject(projectID int64) ([]models.Stage, error) {
	rows, err := s.db.Query(
		"SELECT id, project_id, name, position, created_at, updated_at FROM stages WHERE project_id = ? ORDER BY position",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stages: %v", err)
	}
	defer rows.Close()

	stages := []models.Stage{} // Initialize as empty slice, not nil
	for rows.Next() {
		var stage models.Stage
		err := rows.Scan(&stage.ID, &stage.ProjectID, &stage.Name, &stage.Position, &stage.CreatedAt, &stage.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stage: %v", err)
		}
		stages = append(stages, stage)
	}

	return stages, nil
}

// GetStageByID retrieves a stage by ID
func (s *StageService) GetStageByID(id int64) (*models.Stage, error) {
	var stage models.Stage
	err := s.db.QueryRow(
		"SELECT id, project_id, name, position, created_at, updated_at FROM stages WHERE id = ?",
		id,
	).Scan(&stage.ID, &stage.ProjectID, &stage.Name, &stage.Position, &stage.CreatedAt, &stage.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stage: %v", err)
	}

	return &stage, nil
}

// UpdateStage updates a stage
func (s *StageService) UpdateStage(id int64, name string, position int) (*models.Stage, error) {
	_, err := s.db.Exec(
		"UPDATE stages SET name = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, position, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update stage: %v", err)
	}

	return s.GetStageByID(id)
}

// DeleteStage deletes a stage
func (s *StageService) DeleteStage(id int64) error {
	_, err := s.db.Exec("DELETE FROM stages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete stage: %v", err)
	}
	return nil
}