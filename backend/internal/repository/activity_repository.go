package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"
)

// ActivityRepository handles database operations for activity logs
type ActivityRepository struct {
	db *sql.DB
}

// NewActivityRepository creates a new ActivityRepository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// ActivityLogParams represents parameters for querying activity logs
type ActivityLogParams struct {
	ProjectID int64
	UserID    string
	From      *time.Time
	To        *time.Time
	Page      int
	Limit     int
}

// CreateActivityLog creates a new activity log entry
func (r *ActivityRepository) CreateActivityLog(log *models.ActivityLog) error {
	result, err := r.db.Exec(`
		INSERT INTO activity_logs (project_id, user_id, user_name, action, entity_type, entity_id, description, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.ProjectID,
		log.UserID,
		log.UserName,
		string(log.Action),
		string(log.EntityType),
		log.EntityID,
		log.Description,
		log.Details,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create activity log: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %v", err)
	}
	log.ID = id
	return nil
}

// GetActivityLogsByProject retrieves activity logs for a project with filters
func (r *ActivityRepository) GetActivityLogsByProject(params ActivityLogParams) ([]models.ActivityLogResponse, int64, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "project_id = ?")
	args = append(args, params.ProjectID)

	if params.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, params.UserID)
	}

	if params.From != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *params.From)
	}

	if params.To != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *params.To)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activity_logs WHERE %s", whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count activity logs: %v", err)
	}

	// Apply pagination defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	offset := (params.Page - 1) * params.Limit

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, project_id, user_id, user_name, action, entity_type, entity_id, description, created_at
		FROM activity_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query activity logs: %v", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogResponse
	for rows.Next() {
		var log models.ActivityLogResponse
		var action, entityType string
		var userName sql.NullString
		var entityID sql.NullInt64

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&userName,
			&action,
			&entityType,
			&entityID,
			&log.Description,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity log: %v", err)
		}

		if userName.Valid {
			log.UserName = userName.String
		}
		log.Action = models.ActivityAction(action)
		log.EntityType = models.EntityType(entityType)
		if entityID.Valid {
			log.EntityID = entityID.Int64
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetRecentActivity retrieves recent activity for a project
func (r *ActivityRepository) GetRecentActivity(projectID int64, limit int) ([]models.ActivityLogResponse, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, project_id, user_id, user_name, action, entity_type, entity_id, description, created_at
		FROM activity_logs
		WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent activity: %v", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogResponse
	for rows.Next() {
		var log models.ActivityLogResponse
		var action, entityType string
		var userName sql.NullString
		var entityID sql.NullInt64

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&userName,
			&action,
			&entityType,
			&entityID,
			&log.Description,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %v", err)
		}

		if userName.Valid {
			log.UserName = userName.String
		}
		log.Action = models.ActivityAction(action)
		log.EntityType = models.EntityType(entityType)
		if entityID.Valid {
			log.EntityID = entityID.Int64
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetActivityByEntity retrieves activity for a specific entity
func (r *ActivityRepository) GetActivityByEntity(projectID int64, entityType models.EntityType, entityID int64) ([]models.ActivityLogResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, user_id, user_name, action, entity_type, entity_id, description, created_at
		FROM activity_logs
		WHERE project_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
	`, projectID, string(entityType), entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity activity: %v", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogResponse
	for rows.Next() {
		var log models.ActivityLogResponse
		var action, entityType string
		var userName sql.NullString
		var entID sql.NullInt64

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&userName,
			&action,
			&entityType,
			&entID,
			&log.Description,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %v", err)
		}

		if userName.Valid {
			log.UserName = userName.String
		}
		log.Action = models.ActivityAction(action)
		log.EntityType = models.EntityType(entityType)
		if entID.Valid {
			log.EntityID = entID.Int64
		}

		logs = append(logs, log)
	}

	return logs, nil
}
