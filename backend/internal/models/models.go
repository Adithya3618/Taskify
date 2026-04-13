package models

import "time"

// ProjectMemberRole represents the role of a member in a project
type ProjectMemberRole string

const (
	RoleOwner  ProjectMemberRole = "owner"
	RoleMember ProjectMemberRole = "member"
)

// ProjectMember represents a member of a project
type ProjectMember struct {
	ID        int64             `json:"id"`
	ProjectID int64             `json:"project_id"`
	UserID    string            `json:"user_id"`
	Role      ProjectMemberRole `json:"role"`
	InvitedBy string            `json:"invited_by"`
	JoinedAt  time.Time         `json:"joined_at"`
}

// ProjectMemberResponse is the JSON response for project member data with user info
type ProjectMemberResponse struct {
	ID        int64             `json:"id"`
	ProjectID int64             `json:"project_id"`
	UserID    string            `json:"user_id"`
	UserName  string            `json:"user_name,omitempty"`
	UserEmail string            `json:"user_email,omitempty"`
	Role      ProjectMemberRole `json:"role"`
	InvitedBy string            `json:"invited_by"`
	JoinedAt  time.Time         `json:"joined_at"`
}

// ActivityAction represents the type of activity action
type ActivityAction string

// EntityType represents the type of entity being acted upon
type EntityType string

// Activity action constants
const (
	// Project actions
	ActivityProjectCreated ActivityAction = "project_created"
	ActivityProjectUpdated ActivityAction = "project_updated"
	ActivityProjectDeleted ActivityAction = "project_deleted"

	// Member actions
	ActivityMemberAdded   ActivityAction = "member_added"
	ActivityMemberRemoved ActivityAction = "member_removed"
	ActivityMemberJoined  ActivityAction = "member_joined"

	// Task actions
	ActivityTaskCreated  ActivityAction = "task_created"
	ActivityTaskUpdated  ActivityAction = "task_updated"
	ActivityTaskDeleted  ActivityAction = "task_deleted"
	ActivityTaskAssigned ActivityAction = "task_assigned"
	ActivityTaskMoved    ActivityAction = "task_moved"

	// Label actions
	ActivityLabelAssigned ActivityAction = "label_assigned"
	ActivityLabelRemoved  ActivityAction = "label_removed"

	// Comment actions
	ActivityCommentAdded   ActivityAction = "comment_added"
	ActivityCommentDeleted ActivityAction = "comment_deleted"
)

// Entity types
const (
	EntityProject EntityType = "project"
	EntityTask    EntityType = "task"
	EntityMember  EntityType = "member"
	EntityLabel   EntityType = "label"
	EntityComment EntityType = "comment"
	EntityStage   EntityType = "stage"
)

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID          int64          `json:"id"`
	ProjectID   int64          `json:"project_id"`
	UserID      string         `json:"user_id"`
	UserName    string         `json:"user_name,omitempty"`
	Action      ActivityAction `json:"action"`
	EntityType  EntityType     `json:"entity_type"`
	EntityID    int64          `json:"entity_id,omitempty"`
	Description string         `json:"description"`
	Details     string         `json:"details,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// ActivityLogResponse is the response for activity log
type ActivityLogResponse struct {
	ID          int64          `json:"id"`
	ProjectID   int64          `json:"project_id"`
	UserID      string         `json:"user_id"`
	UserName    string         `json:"user_name,omitempty"`
	Action      ActivityAction `json:"action"`
	EntityType  EntityType     `json:"entity_type"`
	EntityID    int64          `json:"entity_id,omitempty"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Permission constants for project access
const (
	PermissionViewProject   = "view_project"
	PermissionEditProject   = "edit_project"
	PermissionManageStages  = "manage_stages"
	PermissionManageTasks   = "manage_tasks"
	PermissionManageMembers = "manage_members"
	PermissionDeleteProject = "delete_project"
)

// CanEditProject checks if the role can edit project settings
func CanEditProject(role ProjectMemberRole) bool {
	return role == RoleOwner
}

// CanManageMembers checks if the role can manage members
func CanManageMembers(role ProjectMemberRole) bool {
	return role == RoleOwner
}

// CanDeleteProject checks if the role can delete the project
func CanDeleteProject(role ProjectMemberRole) bool {
	return role == RoleOwner
}

// Project represents a project in the system
type Project struct {
	ID          int64     `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Stage represents a stage/column in a project board
type Stage struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task represents a task/card in a stage
type Task struct {
	ID             int64      `json:"id"`
	UserID         string     `json:"user_id"`
	StageID        int64      `json:"stage_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Position       int        `json:"position"`
	Deadline       *time.Time `json:"deadline"`
	Priority       *string    `json:"priority"`
	AssignedTo     *string    `json:"assigned_to"`
	SubtaskCount   int        `json:"subtask_count"`
	CompletedCount int        `json:"completed_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Comment represents a task comment.
type Comment struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	UserID     string    `json:"user_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Subtask represents a task checklist item/subtask.
type Subtask struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	Title       string    `json:"title"`
	IsCompleted bool      `json:"is_completed"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Message represents a chat message in a project
type Message struct {
	ID         int64     `json:"id"`
	UserID     string    `json:"user_id"`
	ProjectID  int64     `json:"project_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}
