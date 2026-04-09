package services

import (
	"database/sql"
	"fmt"

	"backend/internal/auth/repository"
	"backend/internal/models"
	pmrepository "backend/internal/repository"
)

// ProjectMemberService handles business logic for project members
type ProjectMemberService struct {
	db       *sql.DB
	pmRepo   *pmrepository.ProjectMemberRepository
	userRepo *repository.UserRepository
}

// NewProjectMemberService creates a new ProjectMemberService
func NewProjectMemberService(db *sql.DB) *ProjectMemberService {
	return &ProjectMemberService{
		db:       db,
		pmRepo:   pmrepository.NewProjectMemberRepository(db),
		userRepo: repository.NewUserRepository(db),
	}
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

// GetMembersResult represents the result of GetMembers
type GetMembersResult struct {
	Members []models.ProjectMemberResponse
	Page    int
	Limit   int
	Total   int64
}

// GetMembers retrieves all members of a project with their user info
func (s *ProjectMemberService) GetMembers(projectID int64, requesterID string, page, limit int) (*GetMembersResult, error) {
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

	// Apply pagination defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get members with user info (paginated)
	members, total, err := s.pmRepo.GetMembersWithUserInfoPaginated(projectID, page, limit)
	if err != nil {
		return nil, err
	}

	// Return empty array if no members found
	if members == nil {
		members = []models.ProjectMemberResponse{}
	}

	return &GetMembersResult{
		Members: members,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
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
