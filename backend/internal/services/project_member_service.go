package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/auth/repository"
	"backend/internal/models"
	pmrepository "backend/internal/repository"

	"github.com/google/uuid"
)

// ProjectMemberService handles business logic for project members
type ProjectMemberService struct {
	db         *sql.DB
	pmRepo     *pmrepository.ProjectMemberRepository
	inviteRepo *pmrepository.InviteRepository
	userRepo   *repository.UserRepository
}

// NewProjectMemberService creates a new ProjectMemberService
func NewProjectMemberService(db *sql.DB) *ProjectMemberService {
	return &ProjectMemberService{
		db:         db,
		pmRepo:     pmrepository.NewProjectMemberRepository(db),
		inviteRepo: pmrepository.NewInviteRepository(db),
		userRepo:   repository.NewUserRepository(db),
	}
}

// CreateInvite creates an invite link for a project
func (s *ProjectMemberService) CreateInvite(projectID int64, invitedBy string, expiresInHours int) (*models.ProjectInvite, error) {
	// Check if user is owner
	isOwner, err := s.pmRepo.IsOwner(projectID, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check ownership: %v", err)
	}
	if !isOwner {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "only owner can create invites"}
	}

	// Check project exists
	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	// Generate invite ID
	inviteID := uuid.New().String()

	// Calculate expiry time
	var expiresAt time.Time
	if expiresInHours > 0 {
		expiresAt = time.Now().Add(time.Duration(expiresInHours) * time.Hour)
	}

	invite := &models.ProjectInvite{
		ID:        inviteID,
		ProjectID: projectID,
		InvitedBy: invitedBy,
		Role:      "member",
		Status:    "pending",
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	err = s.inviteRepo.CreateInvite(invite)
	if err != nil {
		return nil, err
	}

	return invite, nil
}

// AcceptInviteByID accepts an invite using the invite ID
func (s *ProjectMemberService) AcceptInviteByID(inviteID string, userID string) (*models.ProjectInvite, error) {
	// Get invite
	invite, err := s.inviteRepo.GetInviteByID(inviteID)
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, &ServiceError{Code: "INVITE_NOT_FOUND", Message: "invite not found"}
	}

	// Check if invite is valid
	valid, err := s.inviteRepo.IsValidInvite(inviteID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, &ServiceError{Code: "INVITE_INVALID", Message: "invite is invalid or expired"}
	}

	// Check if user is already a member
	isMember, err := s.pmRepo.IsMember(invite.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, &ServiceError{Code: "ALREADY_MEMBER", Message: "you are already a member of this project"}
	}

	// Add user as member
	_, err = s.pmRepo.AddMember(invite.ProjectID, userID, invite.InvitedBy, invite.Role)
	if err != nil {
		return nil, err
	}

	// Mark invite as accepted
	err = s.inviteRepo.AcceptInvite(inviteID, userID)
	if err != nil {
		return nil, err
	}

	invite.Status = "accepted"
	invite.AcceptedBy = userID
	return invite, nil
}

// GetInvite retrieves an invite by ID
func (s *ProjectMemberService) GetInvite(inviteID string) (*models.ProjectInvite, error) {
	invite, err := s.inviteRepo.GetInviteByID(inviteID)
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, &ServiceError{Code: "INVITE_NOT_FOUND", Message: "invite not found"}
	}
	return invite, nil
}

// projectExists checks if a project exists
func (s *ProjectMemberService) projectExists(projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check project: %v", err)
	}
	return count > 0, nil
}

// AddMember adds a new member to a project
func (s *ProjectMemberService) AddMember(projectID int64, newMemberID, invitedBy string) (*models.ProjectMember, error) {
	// Validate project exists
	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	// Validate user exists
	user, err := s.userRepo.GetUserByID(newMemberID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user: %v", err)
	}
	if user == nil {
		return nil, &ServiceError{Code: "USER_NOT_FOUND", Message: "user not found"}
	}

	// Check if inviter has permission (must be owner)
	isOwner, err := s.pmRepo.IsOwner(projectID, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isOwner {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "only owner can add members"}
	}

	// Check if trying to add self
	if newMemberID == invitedBy {
		return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "cannot add yourself as a member"}
	}

	// Add member
	member, err := s.pmRepo.AddMember(projectID, newMemberID, string(models.RoleMember), invitedBy)
	if err != nil {
		return nil, err
	}

	return member, nil
}

// RemoveMember removes a member from a project
func (s *ProjectMemberService) RemoveMember(projectID int64, targetUserID, removedBy string) error {
	// Validate project exists
	exists, err := s.projectExists(projectID)
	if err != nil {
		return err
	}
	if !exists {
		return &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	// Check if remover has permission (must be owner)
	isOwner, err := s.pmRepo.IsOwner(projectID, removedBy)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isOwner {
		return &ServiceError{Code: "ACCESS_DENIED", Message: "only owner can remove members"}
	}

	// Check if trying to remove self (owner)
	if targetUserID == removedBy {
		return &ServiceError{Code: "INVALID_REQUEST", Message: "cannot remove yourself as owner; transfer ownership first"}
	}

	// Remove member
	err = s.pmRepo.RemoveMember(projectID, targetUserID, removedBy)
	if err != nil {
		return err
	}

	return nil
}

// GetMembers retrieves all members of a project with their user info
// Returns simple array: [{user_id, name, email, role}] - no pagination
func (s *ProjectMemberService) GetMembers(projectID int64, requesterID string) ([]models.ProjectMemberAPIResponse, error) {
	// Validate project exists
	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	// Check if requester has access (owner or member)
	hasAccess, err := s.pmRepo.IsMember(projectID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied"}
	}

	// Get members with user info (no pagination - simple array)
	members, err := s.pmRepo.GetMembersWithUserInfo(projectID)
	if err != nil {
		return nil, err
	}

	// Transform to API response format: [{user_id, name, email, role}]
	result := make([]models.ProjectMemberAPIResponse, 0, len(members))
	for _, m := range members {
		result = append(result, models.ProjectMemberAPIResponse{
			UserID: m.UserID,
			Name:   m.UserName,
			Email:  m.UserEmail,
			Role:   string(m.Role),
		})
	}

	return result, nil
}

// IsMember checks if a user is a member of a project
func (s *ProjectMemberService) IsMember(projectID int64, userID string) (bool, error) {
	return s.pmRepo.IsMember(projectID, userID)
}

// IsOwner checks if a user is the owner of a project
func (s *ProjectMemberService) IsOwner(projectID int64, userID string) (bool, error) {
	return s.pmRepo.IsOwner(projectID, userID)
}

// HasAccess checks if a user has any access to a project
func (s *ProjectMemberService) HasAccess(projectID int64, userID string) (bool, error) {
	// Check if user is owner or member
	isOwner, err := s.pmRepo.IsOwner(projectID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	return s.pmRepo.IsMember(projectID, userID)
}

// GetUserRole gets the role of a user in a project
func (s *ProjectMemberService) GetUserRole(projectID int64, userID string) (models.ProjectMemberRole, error) {
	role, err := s.pmRepo.GetMemberRole(projectID, userID)
	if err != nil {
		return "", err
	}
	return models.ProjectMemberRole(role), nil
}

// ServiceError represents a standardized service error
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// IsServiceError checks if an error is a ServiceError and returns it
func IsServiceError(err error) (*ServiceError, bool) {
	if se, ok := err.(*ServiceError); ok {
		return se, true
	}
	return nil, false
}
