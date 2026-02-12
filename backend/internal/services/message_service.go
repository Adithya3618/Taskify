package services

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
)

type MessageService struct {
	db *sql.DB
}

func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{db: db}
}

// CreateMessage creates a new chat message
func (s *MessageService) CreateMessage(projectID int64, senderName, content string) (*models.Message, error) {
	result, err := s.db.Exec(
		"INSERT INTO messages (project_id, sender_name, content) VALUES (?, ?, ?)",
		projectID, senderName, content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return &models.Message{
		ID:         id,
		ProjectID:  projectID,
		SenderName: senderName,
		Content:    content,
		CreatedAt:  time.Now(),
	}, nil
}

// GetMessagesByProject retrieves all messages for a project
func (s *MessageService) GetMessagesByProject(projectID int64) ([]models.Message, error) {
	rows, err := s.db.Query(
		"SELECT id, project_id, sender_name, content, created_at FROM messages WHERE project_id = ? ORDER BY created_at ASC",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.ProjectID, &message.SenderName, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages for a project (limit 50)
func (s *MessageService) GetRecentMessages(projectID int64) ([]models.Message, error) {
	rows, err := s.db.Query(
		"SELECT id, project_id, sender_name, content, created_at FROM messages WHERE project_id = ? ORDER BY created_at DESC LIMIT 50",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.ProjectID, &message.SenderName, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, message)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(id int64) error {
	_, err := s.db.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}
	return nil
}