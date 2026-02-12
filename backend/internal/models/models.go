package models

import "time"

// Project represents a project in the system
type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Stage represents a stage/column in a project board
type Stage struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task represents a task/card in a stage
type Task struct {
	ID          int64     `json:"id"`
	StageID     int64     `json:"stage_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Message represents a chat message in a project
type Message struct {
	ID         int64     `json:"id"`
	ProjectID  int64     `json:"project_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}