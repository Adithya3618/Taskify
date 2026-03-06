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

// verifyProjectOwnership checks if project belongs to user
func (s *MessageService) verifyProjectOwnership(userID string, projectID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND user_id = ?", projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateMessage creates a new chat message (validates ownership)
func (s *MessageService) CreateMessage(userID string, projectID int64, senderName, content string) (*models.Message, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	result, err := s.db.Exec(
		"INSERT INTO messages (user_id, project_id, sender_name, content) VALUES (?, ?, ?, ?)",
		userID, projectID, senderName, content,
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
		UserID:     userID,
		ProjectID:  projectID,
		SenderName: senderName,
		Content:    content,
		CreatedAt:  time.Now(),
	}, nil
}

// GetMessagesByProject retrieves all messages for a project (validates ownership)
func (s *MessageService) GetMessagesByProject(userID string, projectID int64) ([]models.Message, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	rows, err := s.db.Query(
		"SELECT id, user_id, project_id, sender_name, content, created_at FROM messages WHERE project_id = ? AND user_id = ? ORDER BY created_at ASC",
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.UserID, &message.ProjectID, &message.SenderName, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages for a project (validates ownership)
func (s *MessageService) GetRecentMessages(userID string, projectID int64) ([]models.Message, error) {
	// Verify project belongs to user
	owned, err := s.verifyProjectOwnership(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %v", err)
	}
	if !owned {
		return nil, fmt.Errorf("project not found or access denied")
	}

	rows, err := s.db.Query(
		"SELECT id, user_id, project_id, sender_name, content, created_at FROM messages WHERE project_id = ? AND user_id = ? ORDER BY created_at DESC LIMIT 50",
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.UserID, &message.ProjectID, &message.SenderName, &message.Content, &message.CreatedAt)
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

// DeleteMessage deletes a message (validates ownership)
func (s *MessageService) DeleteMessage(userID string, id int64) error {
	result, err := s.db.Exec("DELETE FROM messages WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found or access denied")
	}

	return nil
}
