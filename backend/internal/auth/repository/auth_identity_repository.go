package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/auth/models"

	"github.com/google/uuid"
)

const googleProvider = "google"

// AuthIdentityRepository handles persistence for external auth providers.
type AuthIdentityRepository struct {
	db *sql.DB
}

// NewAuthIdentityRepository creates a new AuthIdentityRepository.
func NewAuthIdentityRepository(db *sql.DB) *AuthIdentityRepository {
	return &AuthIdentityRepository{db: db}
}

// InitTable creates the auth identities table if it does not exist.
func (r *AuthIdentityRepository) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS auth_identities (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			provider_user_id TEXT NOT NULL,
			provider_email TEXT NOT NULL,
			picture_url TEXT,
			refresh_token TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider, provider_user_id),
			UNIQUE(user_id, provider),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`
	if _, err := r.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create auth identities table: %v", err)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_auth_identities_user_provider ON auth_identities(user_id, provider)",
		"CREATE INDEX IF NOT EXISTS idx_auth_identities_provider_subject ON auth_identities(provider, provider_user_id)",
	}

	for _, indexQuery := range indexes {
		if _, err := r.db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create auth identity index: %v", err)
		}
	}

	return nil
}

// GetByProviderUserID retrieves an auth identity by provider subject.
func (r *AuthIdentityRepository) GetByProviderUserID(provider, providerUserID string) (*models.AuthIdentity, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, provider_email, picture_url, refresh_token, created_at, updated_at
		FROM auth_identities
		WHERE provider = ? AND provider_user_id = ?
	`

	var identity models.AuthIdentity
	var pictureURL sql.NullString
	var refreshToken sql.NullString

	err := r.db.QueryRow(query, provider, providerUserID).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&identity.ProviderEmail,
		&pictureURL,
		&refreshToken,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth identity by provider user id: %v", err)
	}

	identity.PictureURL = nullableStringValue(pictureURL)
	identity.RefreshToken = nullableStringValue(refreshToken)
	return &identity, nil
}

// GetByUserIDAndProvider retrieves an auth identity by user and provider.
func (r *AuthIdentityRepository) GetByUserIDAndProvider(userID, provider string) (*models.AuthIdentity, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, provider_email, picture_url, refresh_token, created_at, updated_at
		FROM auth_identities
		WHERE user_id = ? AND provider = ?
	`

	var identity models.AuthIdentity
	var pictureURL sql.NullString
	var refreshToken sql.NullString

	err := r.db.QueryRow(query, userID, provider).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&identity.ProviderEmail,
		&pictureURL,
		&refreshToken,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth identity by user and provider: %v", err)
	}

	identity.PictureURL = nullableStringValue(pictureURL)
	identity.RefreshToken = nullableStringValue(refreshToken)
	return &identity, nil
}

// UpsertGoogleIdentity creates or updates the google auth identity for a user.
func (r *AuthIdentityRepository) UpsertGoogleIdentity(userID, providerUserID, providerEmail, pictureURL, refreshToken string) error {
	existing, err := r.GetByProviderUserID(googleProvider, providerUserID)
	if err != nil {
		return err
	}

	now := time.Now()
	if existing == nil {
		query := `
			INSERT INTO auth_identities (
				id, user_id, provider, provider_user_id, provider_email, picture_url, refresh_token, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		if _, err := r.db.Exec(
			query,
			uuid.New().String(),
			userID,
			googleProvider,
			providerUserID,
			providerEmail,
			nullableStringArg(pictureURL),
			nullableStringArg(refreshToken),
			now,
			now,
		); err != nil {
			return fmt.Errorf("failed to create google auth identity: %v", err)
		}
		return nil
	}

	updateQuery := `
		UPDATE auth_identities
		SET user_id = ?, provider_email = ?, picture_url = ?, updated_at = ?
		WHERE id = ?
	`
	if _, err := r.db.Exec(updateQuery, userID, providerEmail, nullableStringArg(pictureURL), now, existing.ID); err != nil {
		return fmt.Errorf("failed to update google auth identity: %v", err)
	}

	if refreshToken != "" {
		if err := r.UpdateRefreshToken(existing.ID, refreshToken); err != nil {
			return err
		}
	}

	return nil
}

// UpdateRefreshToken updates the stored refresh token for an auth identity.
func (r *AuthIdentityRepository) UpdateRefreshToken(identityID, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	if _, err := r.db.Exec(
		"UPDATE auth_identities SET refresh_token = ?, updated_at = ? WHERE id = ?",
		refreshToken,
		time.Now(),
		identityID,
	); err != nil {
		return fmt.Errorf("failed to update refresh token: %v", err)
	}

	return nil
}

func nullableStringArg(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullableStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}
