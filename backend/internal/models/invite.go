package models

import "time"

// ProjectInvite represents an invitation to join a project
type ProjectInvite struct {
	ID          string    `json:"id"`
	ProjectID   int64     `json:"project_id"`
	ProjectName string    `json:"project_name,omitempty"`
	InvitedBy   string    `json:"invited_by"`
	Role        string    `json:"role"`   // "member" by default
	Status      string    `json:"status"` // "pending", "accepted", "expired"
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	AcceptedBy  string    `json:"accepted_by,omitempty"`
}
