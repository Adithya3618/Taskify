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

// verifyProjectOwnership checks if project belongs to user
func (s *StageService) verifyProjectOwnership(userID string, projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND user_id = ?", projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateStage creates a new stage for a project (validates ownership)
func (s *StageService) CreateStage(userID string, projectID int64, name string, position int) (*models.Stage, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	result, err := s.db.Exec(
		"INSERT INTO stages (user_id, project_id, name, position) VALUES (?, ?, ?, ?)",
		userID, projectID, name, position,
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
		UserID:    userID,
		ProjectID: projectID,
		Name:      name,
		Position:  position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetStagesByProject retrieves all stages for a project (validates ownership)
func (s *StageService) GetStagesByProject(userID string, projectID int64) ([]models.Stage, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	rows, err := s.db.Query(
		"SELECT id, user_id, project_id, name, position, created_at, updated_at FROM stages WHERE project_id = ? AND user_id = ? ORDER BY position",
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stages: %v", err)
	}
	defer rows.Close()

	stages := []models.Stage{}
	for rows.Next() {
		var stage models.Stage
		err := rows.Scan(&stage.ID, &stage.UserID, &stage.ProjectID, &stage.Name, &stage.Position, &stage.CreatedAt, &stage.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stage: %v", err)
		}
		stages = append(stages, stage)
	}

	return stages, nil
}

// GetStageByID retrieves a stage by ID (validates ownership)
func (s *StageService) GetStageByID(userID string, id int64) (*models.Stage, error) {
	var stage models.Stage
	err := s.db.QueryRow(
		"SELECT id, user_id, project_id, name, position, created_at, updated_at FROM stages WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&stage.ID, &stage.UserID, &stage.ProjectID, &stage.Name, &stage.Position, &stage.CreatedAt, &stage.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stage: %v", err)
	}

	return &stage, nil
}

// UpdateStage updates a stage (validates ownership)
func (s *StageService) UpdateStage(userID string, id int64, name string, position int) (*models.Stage, error) {
	_, err := s.db.Exec(
		"UPDATE stages SET name = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		name, position, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update stage: %v", err)
	}

	return s.GetStageByID(userID, id)
}

// DeleteStage deletes a stage (validates ownership)
func (s *StageService) DeleteStage(userID string, id int64) error {
	result, err := s.db.Exec("DELETE FROM stages WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete stage: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("stage not found or access denied")
	}

	return nil
}
