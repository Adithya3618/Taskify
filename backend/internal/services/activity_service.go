package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

// ActivityService handles business logic for activity logs
type ActivityService struct {
	db            *sql.DB
	activityRepo  *repository.ActivityRepository
	memberService *ProjectMemberService
}

// NewActivityService creates a new ActivityService
func NewActivityService(db *sql.DB, memberService *ProjectMemberService) *ActivityService {
	return &ActivityService{
		db:            db,
		activityRepo:  repository.NewActivityRepository(db),
		memberService: memberService,
	}
}

// LogActivity is the central logging function - all activity logging goes through here
func (s *ActivityService) LogActivity(
	projectID int64,
	userID string,
	userName string,
	action models.ActivityAction,
	entityType models.EntityType,
	entityID int64,
	description string,
	details string,
) error {
	log := &models.ActivityLog{
		ProjectID:   projectID,
		UserID:      userID,
		UserName:    userName,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		Details:     details,
		CreatedAt:   time.Now(),
	}

	err := s.activityRepo.CreateActivityLog(log)
	if err != nil {
		// Log error but don't fail the main operation
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	return nil
}

// GetProjectActivity retrieves activity logs for a project
func (s *ActivityService) GetProjectActivity(
	projectID int64,
	requesterID string,
	params repository.ActivityLogParams,
) ([]models.ActivityLogResponse, int64, error) {
	exists, err := s.projectExists(projectID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check project: %v", err)
	}
	if !exists {
		return nil, 0, &ServiceError{Code: "PROJECT_NOT_FOUND", Message: "project not found"}
	}

	hasAccess, err := s.memberService.HasAccess(projectID, requesterID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, 0, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied"}
	}

	// Set project ID in params
	params.ProjectID = projectID

	// Get activity logs
	logs, total, err := s.activityRepo.GetActivityLogsByProject(params)
	if err != nil {
		return nil, 0, err
	}

	// Return empty array if no logs
	if logs == nil {
		logs = []models.ActivityLogResponse{}
	}

	return logs, total, nil
}

func (s *ActivityService) projectExists(projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRecentActivity retrieves recent activity for a project
func (s *ActivityService) GetRecentActivity(projectID int64, requesterID string, limit int) ([]models.ActivityLogResponse, error) {
	// Validate user has access
	hasAccess, err := s.memberService.HasAccess(projectID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %v", err)
	}
	if !hasAccess {
		return nil, &ServiceError{Code: "ACCESS_DENIED", Message: "access denied"}
	}

	return s.activityRepo.GetRecentActivity(projectID, limit)
}

// Activity logging helpers for common actions

// LogMemberAdded logs when a member is added
func (s *ActivityService) LogMemberAdded(projectID int64, actorID, actorName, newMemberID, newMemberName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityMemberAdded,
		models.EntityMember,
		0,
		fmt.Sprintf("%s added %s to the project", actorName, newMemberName),
		fmt.Sprintf(`{"new_member_id": "%s", "new_member_name": "%s"}`, newMemberID, newMemberName),
	)
}

// LogMemberRemoved logs when a member is removed
func (s *ActivityService) LogMemberRemoved(projectID int64, actorID, actorName, removedMemberID, removedMemberName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityMemberRemoved,
		models.EntityMember,
		0,
		fmt.Sprintf("%s removed %s from the project", actorName, removedMemberName),
		fmt.Sprintf(`{"removed_member_id": "%s", "removed_member_name": "%s"}`, removedMemberID, removedMemberName),
	)
}

// LogTaskCreated logs when a task is created
func (s *ActivityService) LogTaskCreated(projectID int64, actorID, actorName string, taskID int64, taskTitle string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityTaskCreated,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s created task '%s'", actorName, taskTitle),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s"}`, taskID, taskTitle),
	)
}

// LogTaskUpdated logs when a task is updated
func (s *ActivityService) LogTaskUpdated(projectID int64, actorID, actorName string, taskID int64, taskTitle string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityTaskUpdated,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s updated task '%s'", actorName, taskTitle),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s"}`, taskID, taskTitle),
	)
}

// LogTaskDeleted logs when a task is deleted
func (s *ActivityService) LogTaskDeleted(projectID int64, actorID, actorName string, taskTitle string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityTaskDeleted,
		models.EntityTask,
		0,
		fmt.Sprintf("%s deleted task '%s'", actorName, taskTitle),
		fmt.Sprintf(`{"task_title": "%s"}`, taskTitle),
	)
}

// LogTaskMoved logs when a task is moved
func (s *ActivityService) LogTaskMoved(projectID int64, actorID, actorName string, taskID int64, taskTitle, fromStage, toStage string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityTaskMoved,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s moved '%s' from '%s' to '%s'", actorName, taskTitle, fromStage, toStage),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s", "from": "%s", "to": "%s"}`, taskID, taskTitle, fromStage, toStage),
	)
}

// LogLabelCreated logs when a label is created
func (s *ActivityService) LogLabelCreated(projectID int64, actorID, actorName string, labelID int64, labelName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityLabelCreated,
		models.EntityLabel,
		labelID,
		fmt.Sprintf("%s created label '%s'", actorName, labelName),
		fmt.Sprintf(`{"label_id": %d, "label_name": "%s"}`, labelID, labelName),
	)
}

// LogLabelDeleted logs when a label is deleted
func (s *ActivityService) LogLabelDeleted(projectID int64, actorID, actorName string, labelID int64, labelName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityLabelDeleted,
		models.EntityLabel,
		labelID,
		fmt.Sprintf("%s deleted label '%s'", actorName, labelName),
		fmt.Sprintf(`{"label_id": %d, "label_name": "%s"}`, labelID, labelName),
	)
}

// LogLabelAssigned logs when a label is assigned to a task
func (s *ActivityService) LogLabelAssigned(projectID int64, actorID, actorName string, taskID int64, taskTitle, labelName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityLabelAssigned,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s assigned label '%s' to '%s'", actorName, labelName, taskTitle),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s", "label": "%s"}`, taskID, taskTitle, labelName),
	)
}

// LogLabelRemoved logs when a label is removed from a task
func (s *ActivityService) LogLabelRemoved(projectID int64, actorID, actorName string, taskID int64, taskTitle, labelName string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityLabelRemoved,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s removed label '%s' from '%s'", actorName, labelName, taskTitle),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s", "label": "%s"}`, taskID, taskTitle, labelName),
	)
}

// LogCommentAdded logs when a comment is added
func (s *ActivityService) LogCommentAdded(projectID int64, actorID, actorName string, taskID int64, taskTitle string) {
	s.LogActivity(
		projectID,
		actorID,
		actorName,
		models.ActivityCommentAdded,
		models.EntityTask,
		taskID,
		fmt.Sprintf("%s commented on '%s'", actorName, taskTitle),
		fmt.Sprintf(`{"task_id": %d, "task_title": "%s"}`, taskID, taskTitle),
	)
}
