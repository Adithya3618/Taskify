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
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND owner_id = ?", projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *StageService) projectExists(projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *StageService) hasProjectAccess(userID string, projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*)
		FROM projects
		LEFT JOIN project_members
			ON project_members.project_id = projects.id
			AND project_members.user_id = ?
		WHERE projects.id = ?
			AND (projects.owner_id = ? OR project_members.user_id IS NOT NULL)`,
		userID,
		projectID,
		userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateStage creates a new stage for a project (validates ownership)
func (s *StageService) CreateStage(userID string, projectID int64, name string, position int, isFinal int) (*models.Stage, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	result, err := s.db.Exec(
		"INSERT INTO stages (user_id, project_id, name, position, is_final) VALUES (?, ?, ?, ?, ?)",
		userID, projectID, name, position, isFinal,
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
		IsFinal:   isFinal,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetStagesByProject retrieves all stages for a project (validates ownership)
func (s *StageService) GetStagesByProject(userID string, projectID int64) ([]models.Stage, error) {
	hasAccess, err := s.hasProjectAccess(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project access: %v", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("project not found or access denied")
	}

	rows, err := s.db.Query(
		"SELECT id, user_id, project_id, name, position, created_at, updated_at FROM stages WHERE project_id = ? ORDER BY position",
		projectID,
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

// ReorderStages updates all stage positions for a project in a single transaction.
func (s *StageService) ReorderStages(userID string, projectID int64, stageIDs []int64) ([]models.Stage, error) {
	if len(stageIDs) == 0 {
		return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "stage_ids is required"}
	}

	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check project: %v", err)
	}
	if !exists {
		return nil, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	hasAccess, err := s.hasProjectAccess(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	seen := make(map[int64]struct{}, len(stageIDs))
	for _, stageID := range stageIDs {
		if stageID <= 0 {
			return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "stage_ids must contain valid stage IDs"}
		}
		if _, exists := seen[stageID]; exists {
			return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "stage_ids must not contain duplicates"}
		}
		seen[stageID] = struct{}{}
	}

	var totalStages int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM stages WHERE project_id = ?", projectID).Scan(&totalStages); err != nil {
		return nil, fmt.Errorf("failed to count project stages: %v", err)
	}
	if len(stageIDs) != totalStages {
		return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "stage_ids must include exactly all stages in the project"}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin stage reorder transaction: %v", err)
	}
	defer tx.Rollback()

	for position, stageID := range stageIDs {
		var count int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM stages WHERE id = ? AND project_id = ?",
			stageID,
			projectID,
		).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to verify stage %d: %v", stageID, err)
		}
		if count != 1 {
			return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "all stage_ids must belong to the project"}
		}

		result, err := tx.Exec(
			"UPDATE stages SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND project_id = ?",
			position,
			stageID,
			projectID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update stage position: %v", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("failed to check stage update result: %v", err)
		}
		if rowsAffected != 1 {
			return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "all stage_ids must belong to the project"}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit stage reorder: %v", err)
	}

	return s.GetStagesByProject(userID, projectID)
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
