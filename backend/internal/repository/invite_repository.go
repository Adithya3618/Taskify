package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

// InviteRepository handles database operations for project invites
type InviteRepository struct {
	db *sql.DB
}

// NewInviteRepository creates a new InviteRepository
func NewInviteRepository(db *sql.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

// CreateInvite creates a new project invite
func (r *InviteRepository) CreateInvite(invite *models.ProjectInvite) error {
	_, err := r.db.Exec(`
		INSERT INTO project_invites (id, project_id, invited_by, role, status, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		invite.ID,
		invite.ProjectID,
		invite.InvitedBy,
		invite.Role,
		invite.Status,
		invite.CreatedAt,
		invite.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create invite: %v", err)
	}
	return nil
}

// GetInviteByID retrieves an invite by ID
func (r *InviteRepository) GetInviteByID(id string) (*models.ProjectInvite, error) {
	row := r.db.QueryRow(`
		SELECT id, project_id, invited_by, role, status, created_at, expires_at, accepted_by
		FROM project_invites
		WHERE id = ?
	`, id)

	var invite models.ProjectInvite
	var expiresAt sql.NullTime
	var acceptedBy sql.NullString

	err := row.Scan(
		&invite.ID,
		&invite.ProjectID,
		&invite.InvitedBy,
		&invite.Role,
		&invite.Status,
		&invite.CreatedAt,
		&expiresAt,
		&acceptedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %v", err)
	}

	if expiresAt.Valid {
		invite.ExpiresAt = expiresAt.Time
	}
	if acceptedBy.Valid {
		invite.AcceptedBy = acceptedBy.String
	}

	return &invite, nil
}

// AcceptInvite marks an invite as accepted
func (r *InviteRepository) AcceptInvite(id, userID string) error {
	_, err := r.db.Exec(`
		UPDATE project_invites
		SET status = 'accepted', accepted_by = ?
		WHERE id = ?
	`, userID, id)
	if err != nil {
		return fmt.Errorf("failed to accept invite: %v", err)
	}
	return nil
}

// DeleteInvite deletes an invite
func (r *InviteRepository) DeleteInvite(id string) error {
	_, err := r.db.Exec("DELETE FROM project_invites WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete invite: %v", err)
	}
	return nil
}

// GetInvitesByProject retrieves all invites for a project
func (r *InviteRepository) GetInvitesByProject(projectID int64) ([]models.ProjectInvite, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, invited_by, role, status, created_at, expires_at, accepted_by
		FROM project_invites
		WHERE project_id = ?
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invites: %v", err)
	}
	defer rows.Close()

	var invites []models.ProjectInvite
	for rows.Next() {
		var invite models.ProjectInvite
		var expiresAt sql.NullTime
		var acceptedBy sql.NullString

		err := rows.Scan(
			&invite.ID,
			&invite.ProjectID,
			&invite.InvitedBy,
			&invite.Role,
			&invite.Status,
			&invite.CreatedAt,
			&expiresAt,
			&acceptedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invite: %v", err)
		}

		if expiresAt.Valid {
			invite.ExpiresAt = expiresAt.Time
		}
		if acceptedBy.Valid {
			invite.AcceptedBy = acceptedBy.String
		}

		invites = append(invites, invite)
	}

	return invites, nil
}

// IsValidInvite checks if an invite is valid (exists, pending, not expired)
func (r *InviteRepository) IsValidInvite(id string) (bool, error) {
	invite, err := r.GetInviteByID(id)
	if err != nil {
		return false, err
	}
	if invite == nil {
		return false, nil
	}
	if invite.Status != "pending" {
		return false, nil
	}
	if !invite.ExpiresAt.IsZero() && time.Now().After(invite.ExpiresAt) {
		return false, nil
	}
	return true, nil
}
