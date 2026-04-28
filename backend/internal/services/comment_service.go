package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"
)

var (
	ErrCommentNotFoundOrAccessDenied = errors.New("comment not found or access denied")
	ErrTaskNotFoundOrAccessDenied    = errors.New("task not found or access denied")
	ErrCommentContentRequired        = errors.New("comment content is required")
)

type CommentService struct {
	db *sql.DB
}

func NewCommentService(db *sql.DB) *CommentService {
	return &CommentService{db: db}
}

func (s *CommentService) CreateComment(userID string, taskID int64, content string) (*models.Comment, error) {
	if _, err := s.verifyTaskOwnership(userID, taskID); err != nil {
		return nil, err
	}

	normalizedContent, err := normalizeCommentContent(content)
	if err != nil {
		return nil, err
	}

	authorName, err := s.getAuthorName(userID)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(
		"INSERT INTO comments (task_id, user_id, author_name, content) VALUES (?, ?, ?, ?)",
		taskID, userID, authorName, normalizedContent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %v", err)
	}

	now := time.Now()
	return &models.Comment{
		ID:         id,
		TaskID:     taskID,
		UserID:     userID,
		AuthorName: authorName,
		Content:    normalizedContent,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (s *CommentService) GetCommentsByTask(userID string, taskID int64) ([]models.Comment, error) {
	if _, err := s.verifyTaskOwnership(userID, taskID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		"SELECT id, task_id, user_id, author_name, content, created_at, updated_at FROM comments WHERE task_id = ? ORDER BY created_at ASC, id ASC",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %v", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.AuthorName, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading comments: %v", err)
	}

	return comments, nil
}

func (s *CommentService) UpdateComment(userID string, commentID int64, content string) (*models.Comment, error) {
	normalizedContent, err := normalizeCommentContent(content)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(
		"UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		normalizedContent, commentID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return nil, ErrCommentNotFoundOrAccessDenied
	}

	return s.GetCommentByID(userID, commentID)
}

func (s *CommentService) GetCommentByID(userID string, commentID int64) (*models.Comment, error) {
	var comment models.Comment
	err := s.db.QueryRow(
		"SELECT id, task_id, user_id, author_name, content, created_at, updated_at FROM comments WHERE id = ? AND user_id = ?",
		commentID, userID,
	).Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.AuthorName, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrCommentNotFoundOrAccessDenied
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}

	return &comment, nil
}

func (s *CommentService) DeleteComment(userID string, commentID int64) error {
	result, err := s.db.Exec("DELETE FROM comments WHERE id = ? AND user_id = ?", commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return ErrCommentNotFoundOrAccessDenied
	}

	return nil
}

func (s *CommentService) verifyTaskOwnership(userID string, taskID int64) (int64, error) {
	var taskIDFound int64
	err := s.db.QueryRow(`
		SELECT tasks.id
		FROM tasks
		JOIN stages ON tasks.stage_id = stages.id
		JOIN projects ON stages.project_id = projects.id
		WHERE tasks.id = ? AND projects.owner_id = ?`,
		taskID, userID,
	).Scan(&taskIDFound)
	if err == sql.ErrNoRows {
		return 0, ErrTaskNotFoundOrAccessDenied
	}
	if err != nil {
		return 0, fmt.Errorf("failed to verify task ownership: %v", err)
	}
	return taskIDFound, nil
}

func (s *CommentService) getAuthorName(userID string) (string, error) {
	var name string
	err := s.db.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user name: %v", err)
	}
	return strings.TrimSpace(name), nil
}

func normalizeCommentContent(content string) (string, error) {
	normalized := strings.TrimSpace(content)
	if normalized == "" {
		return "", ErrCommentContentRequired
	}
	return normalized, nil
}
