package models

import "time"

// AuthIdentity stores third-party authentication provider data for a user.
type AuthIdentity struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	ProviderEmail  string    `json:"provider_email"`
	PictureURL     string    `json:"picture_url"`
	RefreshToken   string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
