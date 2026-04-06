package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"backend/internal/auth/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		user.ID,
		user.Name,
		user.Email,
		nullablePasswordHash(user.PasswordHash),
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by email address
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	row := r.db.QueryRow(query, email)

	var user models.User
	var passwordHash sql.NullString
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&passwordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}
	user.PasswordHash = nullableStringValue(passwordHash)
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	var user models.User
	var passwordHash sql.NullString
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&passwordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	user.PasswordHash = nullableStringValue(passwordHash)
	return &user, nil
}

// EmailExists checks if an email already exists in the database
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %v", err)
	}
	return exists, nil
}

// InitTable creates the users table if it doesn't exist
func (r *UserRepository) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			role TEXT DEFAULT 'user',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if err := r.migratePasswordHashNullable(); err != nil {
		return err
	}

	// Create index on email for faster lookups
	indexQuery := `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`
	_, err = r.db.Exec(indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create users email index: %v", err)
	}

	return nil
}

// SetUserActive updates the active flag for a user.
func (r *UserRepository) SetUserActive(userID string, isActive bool) error {
	if _, err := r.db.Exec(
		"UPDATE users SET is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		isActive,
		userID,
	); err != nil {
		return fmt.Errorf("failed to update user active state: %v", err)
	}
	return nil
}

// UpdatePassword updates a user's password hash by email
func (r *UserRepository) UpdatePassword(email, newPasswordHash string) error {
	result, err := r.db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE email = ?",
		newPasswordHash, email,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) migratePasswordHashNullable() error {
	notNull, err := r.passwordHashIsNotNull()
	if err != nil {
		return err
	}
	if !notNull {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin user migration: %v", err)
	}
	defer tx.Rollback()

	statements := []string{
		"ALTER TABLE users RENAME TO users_legacy",
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			role TEXT DEFAULT 'user',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO users (id, name, email, password_hash, role, is_active, created_at, updated_at)
		 SELECT id, name, email, password_hash, role, is_active, created_at, updated_at FROM users_legacy`,
		"DROP TABLE users_legacy",
	}

	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("failed to migrate users table: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit users migration: %v", err)
	}

	return nil
}

func (r *UserRepository) passwordHashIsNotNull() (bool, error) {
	rows, err := r.db.Query("PRAGMA table_info(users)")
	if err != nil {
		return false, fmt.Errorf("failed to inspect users table: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, fmt.Errorf("failed to scan users schema: %v", err)
		}

		if strings.EqualFold(name, "password_hash") {
			return notNull == 1, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("failed while reading users schema: %v", err)
	}

	return false, nil
}

func nullablePasswordHash(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
