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

// CreateProject creates a new project (owner_id from JWT)
func (s *ProjectService) CreateProject(ownerID, name, description string) (*models.Project, error) {
	result, err := s.db.Exec(
		"INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)",
		ownerID, name, description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	// Automatically add owner as project member
	_, err = s.db.Exec(
		"INSERT INTO project_members (project_id, user_id, role, invited_by, joined_at) VALUES (?, ?, ?, ?, ?)",
		id, ownerID, "owner", ownerID, time.Now(),
	)
	if err != nil {
		// Log but don't fail - project is created
		fmt.Printf("Warning: failed to add owner to project_members: %v\n", err)
	}

	return &models.Project{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetAllProjects retrieves all projects where user is owner or member
func (s *ProjectService) GetAllProjects(userID string) ([]models.Project, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT p.id, p.owner_id, p.name, p.description, p.created_at, p.updated_at 
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.owner_id = ? OR pm.user_id = ?
		ORDER BY p.created_at DESC
	`, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.ID, &project.OwnerID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, project)
	}
	return projects, nil
}

// GetProject retrieves a single project by ID (user must be owner or member)
func (s *ProjectService) GetProject(userID string, id int64) (*models.Project, error) {
	row := s.db.QueryRow(`
		SELECT p.id, p.owner_id, p.name, p.description, p.created_at, p.updated_at 
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.id = ? AND (p.owner_id = ? OR pm.user_id = ?)
	`, id, userID, userID)

	var project models.Project
	err := row.Scan(&project.ID, &project.OwnerID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	return &project, nil
}

// GetProjectByID retrieves a project by ID (no access check - for internal use)
func (s *ProjectService) GetProjectByID(id int64) (*models.Project, error) {
	row := s.db.QueryRow(
		"SELECT id, owner_id, name, description, created_at, updated_at FROM projects WHERE id = ?",
		id)

	var project models.Project
	err := row.Scan(&project.ID, &project.OwnerID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	return &project, nil
}

// IsOwner checks if user is the owner of the project
func (s *ProjectService) IsOwner(userID string, projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM projects WHERE id = ? AND owner_id = ?",
		projectID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %v", err)
	}
	return count > 0, nil
}

// UpdateProject updates a project (must be owner)
func (s *ProjectService) UpdateProject(userID string, id int64, name, description string) (*models.Project, error) {
	// Check if user is owner
	isOwner, err := s.IsOwner(userID, id)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, fmt.Errorf("access denied: only owner can update project")
	}

	_, err = s.db.Exec(
		"UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ? AND owner_id = ?",
		name, description, time.Now(), id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %v", err)
	}

	return s.GetProjectByID(id)
}

// DeleteProject deletes a project (must be owner)
func (s *ProjectService) DeleteProject(userID string, id int64) error {
	// Check if user is owner
	isOwner, err := s.IsOwner(userID, id)
	if err != nil {
		return err
	}
	if !isOwner {
		return fmt.Errorf("access denied: only owner can delete project")
	}

	result, err := s.db.Exec("DELETE FROM projects WHERE id = ? AND owner_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
