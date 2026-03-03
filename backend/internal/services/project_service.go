package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

type ProjectService struct {
	db *sql.DB
}

func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject creates a new project (user_id from JWT)
func (s *ProjectService) CreateProject(userID, name, description string) (*models.Project, error) {
	result, err := s.db.Exec(
		"INSERT INTO projects (user_id, name, description) VALUES (?, ?, ?)",
		userID, name, description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return &models.Project{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetAllProjects retrieves all projects for a specific user
func (s *ProjectService) GetAllProjects(userID string) ([]models.Project, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, name, description, created_at, updated_at FROM projects WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, project)
	}
	return projects, nil
}

// GetProject retrieves a single project by ID (must belong to user)
func (s *ProjectService) GetProject(userID string, id int64) (*models.Project, error) {
	row := s.db.QueryRow(
		"SELECT id, user_id, name, description, created_at, updated_at FROM projects WHERE id = ? AND user_id = ?",
		id, userID,
	)

	var project models.Project
	err := row.Scan(&project.ID, &project.UserID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	return &project, nil
}

// UpdateProject updates a project (must belong to user)
func (s *ProjectService) UpdateProject(userID string, id int64, name, description string) (*models.Project, error) {
	_, err := s.db.Exec(
		"UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ? AND user_id = ?",
		name, description, time.Now(), id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %v", err)
	}

	return s.GetProject(userID, id)
}

// DeleteProject deletes a project (must belong to user)
func (s *ProjectService) DeleteProject(userID string, id int64) error {
	result, err := s.db.Exec("DELETE FROM projects WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found or access denied")
	}

	return nil
}
