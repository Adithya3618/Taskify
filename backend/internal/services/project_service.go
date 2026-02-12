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

// CreateProject creates a new project
func (s *ProjectService) CreateProject(name, description string) (*models.Project, error) {
	result, err := s.db.Exec(
		"INSERT INTO projects (name, description) VALUES (?, ?)",
		name, description,
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
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetAllProjects retrieves all projects
func (s *ProjectService) GetAllProjects() ([]models.Project, error) {
	rows, err := s.db.Query(
		"SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// GetProjectByID retrieves a project by ID
func (s *ProjectService) GetProjectByID(id int64) (*models.Project, error) {
	var project models.Project
	err := s.db.QueryRow(
		"SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?",
		id,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	return &project, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(id int64, name, description string) (*models.Project, error) {
	_, err := s.db.Exec(
		"UPDATE projects SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, description, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %v", err)
	}

	return s.GetProjectByID(id)
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(id int64) error {
	_, err := s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}
	return nil
}