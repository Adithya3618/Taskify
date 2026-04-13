package services

import (
	"database/sql"
	"fmt"

	"backend/internal/models"
	"backend/internal/repository"
)

// TaskLabelService handles business logic for task-label associations
type TaskLabelService struct {
	db            *sql.DB
	taskLabelRepo *repository.TaskLabelRepository
	labelRepo     *repository.LabelRepository
	pmService     *ProjectMemberService
	activitySvc   *ActivityService
}

// MaxLabelsPerTask defines the maximum number of labels that can be assigned to a task
const MaxLabelsPerTask = 10

// NewTaskLabelService creates a new TaskLabelService
func NewTaskLabelService(db *sql.DB, pmService *ProjectMemberService, activitySvc *ActivityService) *TaskLabelService {
	return &TaskLabelService{
		db:            db,
		taskLabelRepo: repository.NewTaskLabelRepository(db),
		labelRepo:     repository.NewLabelRepository(db),
		pmService:     pmService,
		activitySvc:   activitySvc,
	}
}

// AssignLabel assigns a label to a task
func (s *TaskLabelService) AssignLabel(taskID, labelID int64, userID, userName string) error {
	// Get task to find its project
	taskProjectID, err := s.getTaskProjectID(taskID)
	if err != nil {
		return err
	}
	if taskProjectID == 0 {
		return &ServiceError{Code: "TASK_NOT_FOUND", Message: "task not found"}
	}

	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(taskProjectID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	// Check label count limit
	currentCount, err := s.getTaskLabelCount(taskID)
	if err != nil {
		return err
	}
	if currentCount >= MaxLabelsPerTask {
		return &ServiceError{Code: "LABEL_LIMIT_EXCEEDED", Message: fmt.Sprintf("task cannot have more than %d labels", MaxLabelsPerTask)}
	}

	// Get label to validate it exists and belongs to same project
	label, err := s.labelRepo.GetLabelByID(labelID)
	if err != nil {
		return err
	}
	if label == nil {
		return &ServiceError{Code: "LABEL_NOT_FOUND", Message: "label not found"}
	}
	if label.ProjectID != taskProjectID {
		return &ServiceError{Code: "CROSS_PROJECT_LABEL", Message: "cannot assign label from different project"}
	}

	// Get task title for activity log
	taskTitle, _ := s.getTaskTitle(taskID)

	// Assign label
	if err := s.taskLabelRepo.AssignLabelToTask(taskID, labelID); err != nil {
		return err
	}

	// Log activity
	if s.activitySvc != nil {
		s.activitySvc.LogLabelAssigned(taskProjectID, userID, userName, taskID, taskTitle, label.Name)
	}

	return nil
}

// RemoveLabel removes a label from a task
func (s *TaskLabelService) RemoveLabel(taskID, labelID int64, userID, userName string) error {
	// Get task to find its project
	taskProjectID, err := s.getTaskProjectID(taskID)
	if err != nil {
		return err
	}
	if taskProjectID == 0 {
		return &ServiceError{Code: "TASK_NOT_FOUND", Message: "task not found"}
	}

	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(taskProjectID, userID)
	if err != nil {
		return fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	// Get label name for activity log
	label, _ := s.labelRepo.GetLabelByID(labelID)
	labelName := ""
	if label != nil {
		labelName = label.Name
	}

	// Get task title for activity log
	taskTitle, _ := s.getTaskTitle(taskID)

	// Remove label
	if err := s.taskLabelRepo.RemoveLabelFromTask(taskID, labelID); err != nil {
		return err
	}

	// Log activity
	if s.activitySvc != nil {
		s.activitySvc.LogLabelRemoved(taskProjectID, userID, userName, taskID, taskTitle, labelName)
	}

	return nil
}

// GetTaskLabels retrieves all labels for a task
func (s *TaskLabelService) GetTaskLabels(taskID int64, userID string) ([]models.Label, error) {
	// Get task to find its project
	taskProjectID, err := s.getTaskProjectID(taskID)
	if err != nil {
		return nil, err
	}
	if taskProjectID == 0 {
		return nil, &ServiceError{Code: "TASK_NOT_FOUND", Message: "task not found"}
	}

	// Validate user is project member
	hasAccess, err := s.pmService.HasAccess(taskProjectID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied to this project"}
	}

	labels, err := s.taskLabelRepo.GetLabelsByTask(taskID)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = []models.Label{}
	}

	return labels, nil
}

// getTaskProjectID retrieves the project ID for a task
func (s *TaskLabelService) getTaskProjectID(taskID int64) (int64, error) {
	var projectID int64
	err := s.db.QueryRow(`
		SELECT s.project_id
		FROM tasks t
		JOIN stages s ON t.stage_id = s.id
		WHERE t.id = ?
	`, taskID).Scan(&projectID)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get task project: %v", err)
	}
	return projectID, nil
}

// getTaskTitle retrieves the title of a task
func (s *TaskLabelService) getTaskTitle(taskID int64) (string, error) {
	var title string
	err := s.db.QueryRow("SELECT title FROM tasks WHERE id = ?", taskID).Scan(&title)
	if err != nil {
		return "", nil // Ignore error, return empty string
	}
	return title, nil
}

// getTaskLabelCount retrieves the number of labels assigned to a task
func (s *TaskLabelService) getTaskLabelCount(taskID int64) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count task labels: %v", err)
	}
	return count, nil
}
