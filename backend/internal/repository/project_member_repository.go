package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/models"
)

// ProjectMemberRepository handles database operations for project members
type ProjectMemberRepository struct {
	db *sql.DB
}

// NewProjectMemberRepository creates a new ProjectMemberRepository
func NewProjectMemberRepository(db *sql.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

// AddMember adds a new member to a project with transaction safety
func (r *ProjectMemberRepository) AddMember(projectID int64, userID, role, invitedBy string) (*models.ProjectMember, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if member already exists inside the transaction. SQLite does not
	// support SELECT ... FOR UPDATE; the UNIQUE constraint still protects
	// against concurrent duplicate inserts.
	var existingID int64
	err = tx.QueryRow(
		"SELECT id FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			return nil, fmt.Errorf("user is already a member of this project")
		}
		return nil, fmt.Errorf("failed to check existing membership: %v", err)
	}

	// Insert new member
	result, err := tx.Exec(
		"INSERT INTO project_members (project_id, user_id, role, invited_by, joined_at) VALUES (?, ?, ?, ?, ?)",
		projectID, userID, role, invitedBy, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add member: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	// Log activity
	details, _ := json.Marshal(map[string]interface{}{
		"role": role, "invited_by": invitedBy,
	})
	_, err = tx.Exec(
		`INSERT INTO activity_logs
			(project_id, user_id, action, entity_type, description, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		projectID,
		invitedBy,
		models.ActivityMemberAdded,
		models.EntityMember,
		fmt.Sprintf("Added member %s", userID),
		string(details),
		time.Now(),
	)
	if err != nil {
		// Activity logging is non-critical, log but continue
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &models.ProjectMember{
		ID:        id,
		ProjectID: projectID,
		UserID:    userID,
		Role:      models.ProjectMemberRole(role),
		InvitedBy: invitedBy,
		JoinedAt:  time.Now(),
	}, nil
}

// RemoveMember removes a member from a project with transaction safety
func (r *ProjectMemberRepository) RemoveMember(projectID int64, userID, removedBy string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if member exists inside the transaction. SQLite does not support
	// SELECT ... FOR UPDATE.
	var existingRole string
	err = tx.QueryRow(
		"SELECT role FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	).Scan(&existingRole)
	if err == sql.ErrNoRows {
		return fmt.Errorf("member not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check membership: %v", err)
	}

	// Prevent owner removal
	if existingRole == "owner" {
		return fmt.Errorf("cannot remove owner from project")
	}

	// Delete member
	_, err = tx.Exec(
		"DELETE FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove member: %v", err)
	}

	// Log activity
	_, err = tx.Exec(
		`INSERT INTO activity_logs
			(project_id, user_id, action, entity_type, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		projectID,
		removedBy,
		models.ActivityMemberRemoved,
		models.EntityMember,
		fmt.Sprintf("Removed member %s", userID),
		time.Now(),
	)
	if err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetMembers retrieves all members of a project (without pagination)
func (r *ProjectMemberRepository) GetMembers(projectID int64) ([]models.ProjectMember, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, user_id, role, COALESCE(invited_by, ''), joined_at 
		FROM project_members 
		WHERE project_id = ? 
		ORDER BY joined_at ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %v", err)
	}
	defer rows.Close()

	var members []models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		var role string
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &role, &member.InvitedBy, &member.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %v", err)
		}
		member.Role = models.ProjectMemberRole(role)
		members = append(members, member)
	}
	return members, nil
}

// GetMembersPaginated retrieves paginated members of a project
func (r *ProjectMemberRepository) GetMembersPaginated(projectID int64, page, limit int) ([]models.ProjectMember, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM project_members WHERE project_id = ?",
		projectID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count members: %v", err)
	}

	// Apply pagination defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := (page - 1) * limit

	// Get paginated results
	rows, err := r.db.Query(`
		SELECT id, project_id, user_id, role, COALESCE(invited_by, ''), joined_at 
		FROM project_members 
		WHERE project_id = ? 
		ORDER BY joined_at ASC
		LIMIT ? OFFSET ?
	`, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query members: %v", err)
	}
	defer rows.Close()

	var members []models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		var role string
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &role, &member.InvitedBy, &member.JoinedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan member: %v", err)
		}
		member.Role = models.ProjectMemberRole(role)
		members = append(members, member)
	}

	return members, total, nil
}

// GetMembersWithUserInfo retrieves all members with their user information (without pagination)
func (r *ProjectMemberRepository) GetMembersWithUserInfo(projectID int64) ([]models.ProjectMemberResponse, error) {
	rows, err := r.db.Query(`
		SELECT pm.id, pm.project_id, pm.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), pm.role, COALESCE(pm.invited_by, ''), pm.joined_at 
		FROM project_members pm
		LEFT JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = ? 
		ORDER BY pm.joined_at ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %v", err)
	}
	defer rows.Close()

	var members []models.ProjectMemberResponse
	for rows.Next() {
		var member models.ProjectMemberResponse
		var role string
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &member.UserName, &member.UserEmail, &role, &member.InvitedBy, &member.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %v", err)
		}
		member.Role = models.ProjectMemberRole(role)
		members = append(members, member)
	}
	return members, nil
}

// GetMembersWithUserInfoPaginated retrieves paginated members with their user information
func (r *ProjectMemberRepository) GetMembersWithUserInfoPaginated(projectID int64, page, limit int) ([]models.ProjectMemberResponse, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM project_members WHERE project_id = ?",
		projectID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count members: %v", err)
	}

	// Apply pagination defaults
	if page < 1 {
		return r.getMembersWithUserInfoNoPagination(projectID, total)
	}

	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	return r.getMembersWithUserInfoPaginated(projectID, limit, offset, total)
}

func (r *ProjectMemberRepository) getMembersWithUserInfoNoPagination(projectID int64, total int64) ([]models.ProjectMemberResponse, int64, error) {
	rows, err := r.db.Query(`
		SELECT pm.id, pm.project_id, pm.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), pm.role, COALESCE(pm.invited_by, ''), pm.joined_at 
		FROM project_members pm
		LEFT JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = ? 
		ORDER BY pm.joined_at ASC
	`, projectID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query members: %v", err)
	}
	defer rows.Close()

	var members []models.ProjectMemberResponse
	for rows.Next() {
		var member models.ProjectMemberResponse
		var role string
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &member.UserName, &member.UserEmail, &role, &member.InvitedBy, &member.JoinedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan member: %v", err)
		}
		member.Role = models.ProjectMemberRole(role)
		members = append(members, member)
	}

	return members, total, nil
}

func (r *ProjectMemberRepository) getMembersWithUserInfoPaginated(projectID int64, limit, offset int, total int64) ([]models.ProjectMemberResponse, int64, error) {
	rows, err := r.db.Query(`
		SELECT pm.id, pm.project_id, pm.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), pm.role, COALESCE(pm.invited_by, ''), pm.joined_at 
		FROM project_members pm
		LEFT JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = ? 
		ORDER BY pm.joined_at ASC
		LIMIT ? OFFSET ?
	`, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query members: %v", err)
	}
	defer rows.Close()

	var members []models.ProjectMemberResponse
	for rows.Next() {
		var member models.ProjectMemberResponse
		var role string
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &member.UserName, &member.UserEmail, &role, &member.InvitedBy, &member.JoinedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan member: %v", err)
		}
		member.Role = models.ProjectMemberRole(role)
		members = append(members, member)
	}

	return members, total, nil
}

// IsMember checks if a user is a member of a project
func (r *ProjectMemberRepository) IsMember(projectID int64, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %v", err)
	}
	return count > 0, nil
}

// IsOwner checks if a user is the owner of a project
func (r *ProjectMemberRepository) IsOwner(projectID int64, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM project_members WHERE project_id = ? AND user_id = ? AND role = 'owner'",
		projectID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %v", err)
	}
	return count > 0, nil
}

// GetMemberRole gets the role of a user in a project
func (r *ProjectMemberRepository) GetMemberRole(projectID int64, userID string) (string, error) {
	var role string
	err := r.db.QueryRow(
		"SELECT role FROM project_members WHERE project_id = ? AND user_id = ?",
		projectID, userID,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %v", err)
	}
	return role, nil
}

// Exists checks if a membership record exists
func (r *ProjectMemberRepository) Exists(projectID int64, userID string) (bool, error) {
	return r.IsMember(projectID, userID)
}

// GetMemberProjects retrieves all projects a user is a member of
func (r *ProjectMemberRepository) GetMemberProjects(userID string) ([]int64, error) {
	rows, err := r.db.Query(
		"SELECT project_id FROM project_members WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query member projects: %v", err)
	}
	defer rows.Close()

	var projectIDs []int64
	for rows.Next() {
		var projectID int64
		if err := rows.Scan(&projectID); err != nil {
			return nil, fmt.Errorf("failed to scan project id: %v", err)
		}
		projectIDs = append(projectIDs, projectID)
	}

	return projectIDs, nil
}

// EnsureOwnerConsistency ensures projects.owner_id matches the owner in project_members
func (r *ProjectMemberRepository) EnsureOwnerConsistency(projectID int64) error {
	// Get owner from project_members
	var ownerID string
	err := r.db.QueryRow(
		"SELECT user_id FROM project_members WHERE project_id = ? AND role = 'owner'",
		projectID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no owner found in project_members for project %d", projectID)
	}
	if err != nil {
		return fmt.Errorf("failed to get owner from project_members: %v", err)
	}

	// Update projects.owner_id to match
	_, err = r.db.Exec(
		"UPDATE projects SET owner_id = ? WHERE id = ?",
		ownerID, projectID,
	)
	if err != nil {
		return fmt.Errorf("failed to update projects.owner_id: %v", err)
	}

	return nil
}
