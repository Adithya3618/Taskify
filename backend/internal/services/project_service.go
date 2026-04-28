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

// GetProjectStats retrieves project statistics (tasks, completion, etc.)
// Authorization: user must be a member of the project (checked via project_members table)
func (s *ProjectService) GetProjectStats(projectID int64, userID string) (*models.ProjectStats, error) {
	// 1. Check if project exists
	var projectExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&projectExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check project: %v", err)
	}
	if projectExists == 0 {
		return nil, fmt.Errorf("project not found")
	}

	// 2. Check if user is a member
	var isMember int
	err = s.db.QueryRow(
		"SELECT COUNT(*) FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	).Scan(&isMember)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %v", err)
	}
	if isMember == 0 {
		return nil, fmt.Errorf("project not found") // Return 404 per requirement
	}

	// 3. Get total tasks, completed, and overdue in single query
	// Note: Using stages.is_final column for completion tracking
	var totalTasks, completedTasks, overdueTasks int64
	err = s.db.QueryRow(`
		SELECT 
			COUNT(t.id) as total,
			COALESCE(SUM(CASE WHEN st.is_final = 1 THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN t.deadline IS NOT NULL AND t.deadline < CURRENT_TIMESTAMP AND st.is_final != 1 THEN 1 ELSE 0 END), 0) as overdue
		FROM tasks t
		JOIN stages st ON t.stage_id = st.id
		WHERE st.project_id = ?`,
		projectID,
	).Scan(&totalTasks, &completedTasks, &overdueTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %v", err)
	}

	// 4. Get tasks by stage (including stages with 0 tasks)
	rows, err := s.db.Query(`
		SELECT s.id, s.name, COALESCE(COUNT(t.id), 0) as count
		FROM stages s
		LEFT JOIN tasks t ON t.stage_id = s.id
		WHERE s.project_id = ?
		GROUP BY s.id, s.name
		ORDER BY s.position`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by stage: %v", err)
	}
	defer rows.Close()

	tasksByStage := make([]models.StageTaskCount, 0)
	for rows.Next() {
		var stage models.StageTaskCount
		if err := rows.Scan(&stage.StageID, &stage.StageName, &stage.Count); err != nil {
			return nil, fmt.Errorf("failed to scan stage: %v", err)
		}
		tasksByStage = append(tasksByStage, stage)
	}

	// 5. Calculate completion rate
	var completionRate float64
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	return &models.ProjectStats{
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		OverdueTasks:   overdueTasks,
		TasksByStage:   tasksByStage,
		CompletionRate: completionRate,
	}, nil
}
