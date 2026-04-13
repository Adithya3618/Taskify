package services

import (
	"database/sql"
	"fmt"
	"regexp"

	"backend/internal/models"
	"backend/internal/repository"
)

// LabelService handles business logic for labels
type LabelService struct {
	db          *sql.DB
	labelRepo   *repository.LabelRepository
	pmService   *ProjectMemberService
	activitySvc *ActivityService
}

// colorRegex validates hex color format (#RRGGBB)
var colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// NewLabelService creates a new LabelService
func NewLabelService(db *sql.DB, pmService *ProjectMemberService, activitySvc *ActivityService) *LabelService {
	return &LabelService{
		db:          db,
		labelRepo:   repository.NewLabelRepository(db),
		pmService:   pmService,
		activitySvc: activitySvc,
	}
}

// CreateLabel creates a new label in a project
func (s *LabelService) CreateLabel(projectID int64, name, color, userID, userName string) (*models.Label, error) {
	// Validate project exists
	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	// Validate and set default color
	if color == "" {
		color = "#808080"
	} else if !colorRegex.MatchString(color) {
		return nil, &ServiceError{Code: "INVALID_COLOR", Message: "color must be valid hex format (e.g., #FF5733)"}
	}

	// Create label
	label, err := s.labelRepo.CreateLabel(projectID, name, color, userID)
	if err != nil {
		return nil, err
	}

	// Log activity
	if s.activitySvc != nil {
		s.activitySvc.LogLabelCreated(projectID, userID, userName, label.ID, name)
	}

	return label, nil
}

// GetProjectLabels retrieves all labels for a project
func (s *LabelService) GetProjectLabels(projectID int64, userID string) ([]models.Label, error) {
	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	labels, err := s.labelRepo.GetLabelsByProject(projectID)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = []models.Label{}
	}

	return labels, nil
}

// DeleteLabel deletes a label
func (s *LabelService) DeleteLabel(labelID int64, userID, userName string) error {
	// Get label to find project ID
	label, err := s.labelRepo.GetLabelByID(labelID)
	if err != nil {
		return err
	}
	if label == nil {
		return &ServiceError{Code: "LABEL_NOT_FOUND", Message: "label not found"}
	}

	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(label.ProjectID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	// Check permission: only creator or project owner can delete
	isOwner, err := s.pmService.IsOwner(label.ProjectID, userID)
	if err != nil {
		return fmt.Errorf("failed to check ownership: %v", err)
	}
	isCreator := label.CreatedBy == userID

	if !isOwner && !isCreator {
		return &ServiceError{Code: "PERMISSION_DENIED", Message: "only the label creator or project owner can delete this label"}
	}

	// Delete label (cascade will remove task associations)
	if err := s.labelRepo.DeleteLabel(labelID); err != nil {
		return err
	}

	// Log activity
	if s.activitySvc != nil {
		s.activitySvc.LogLabelDeleted(label.ProjectID, userID, userName, labelID, label.Name)
	}

	return nil
}

// GetLabelByID retrieves a label by ID
func (s *LabelService) GetLabelByID(labelID int64) (*models.Label, error) {
	return s.labelRepo.GetLabelByID(labelID)
}

// projectExists checks if a project exists
func (s *LabelService) projectExists(projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check project: %v", err)
	}
	return count > 0, nil
}
