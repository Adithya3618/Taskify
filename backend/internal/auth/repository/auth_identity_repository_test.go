package repository

import (
	"database/sql"
	"testing"
	"time"

	"backend/internal/auth/models"

	_ "github.com/mattn/go-sqlite3"
)

func newTestAuthDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	userRepo := NewUserRepository(db)
	if err := userRepo.InitTable(); err != nil {
		t.Fatalf("failed to init users table: %v", err)
	}

	identityRepo := NewAuthIdentityRepository(db)
	if err := identityRepo.InitTable(); err != nil {
		t.Fatalf("failed to init auth identities table: %v", err)
	}

	return db
}

func TestAuthIdentityRepositoryUpsertGoogleIdentity(t *testing.T) {
	db := newTestAuthDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	identityRepo := NewAuthIdentityRepository(db)

	user := &models.User{
		ID:        "user-1",
		Name:      "Jane",
		Email:     "jane@example.com",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := userRepo.CreateUser(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := identityRepo.UpsertGoogleIdentity(user.ID, "google-sub-1", user.Email, "https://img.example.com/jane.png", "refresh-1"); err != nil {
		t.Fatalf("failed to upsert google identity: %v", err)
	}

	identity, err := identityRepo.GetByProviderUserID("google", "google-sub-1")
	if err != nil {
		t.Fatalf("failed to fetch identity: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity to exist")
	}
	if identity.UserID != user.ID {
		t.Fatalf("expected user id %q, got %q", user.ID, identity.UserID)
	}
	if identity.RefreshToken != "refresh-1" {
		t.Fatalf("expected refresh token to be stored")
	}

	if err := identityRepo.UpsertGoogleIdentity(user.ID, "google-sub-1", user.Email, "https://img.example.com/new.png", "refresh-2"); err != nil {
		t.Fatalf("failed to update google identity: %v", err)
	}

	updatedIdentity, err := identityRepo.GetByUserIDAndProvider(user.ID, "google")
	if err != nil {
		t.Fatalf("failed to fetch updated identity: %v", err)
	}
	if updatedIdentity.PictureURL != "https://img.example.com/new.png" {
		t.Fatalf("expected picture url to update, got %q", updatedIdentity.PictureURL)
	}
	if updatedIdentity.RefreshToken != "refresh-2" {
		t.Fatalf("expected refresh token to update, got %q", updatedIdentity.RefreshToken)
	}
}
