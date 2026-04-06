package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"backend/internal/auth/repository"
	_ "github.com/mattn/go-sqlite3"
)

func newAuthServiceForTest(t *testing.T) (*AuthService, *repository.AuthIdentityRepository, *repository.UserRepository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userRepo := repository.NewUserRepository(db)
	if err := userRepo.InitTable(); err != nil {
		t.Fatalf("failed to init users table: %v", err)
	}

	identityRepo := repository.NewAuthIdentityRepository(db)
	if err := identityRepo.InitTable(); err != nil {
		t.Fatalf("failed to init identity table: %v", err)
	}

	service := NewAuthService(
		userRepo,
		identityRepo,
		NewJWTService("test-secret", 24),
		NewOTPService(),
		&EmailService{},
		&GoogleAuthService{},
		NewOAuthStateService(10*time.Minute),
	)

	return service, identityRepo, userRepo
}

func TestGoogleLoginLinksExistingEmailUser(t *testing.T) {
	service, identityRepo, _ := newAuthServiceForTest(t)

	if _, err := service.Register(RegisterRequest{
		Name:     "Existing User",
		Email:    "existing@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("failed to register local user: %v", err)
	}

	resp, err := service.completeGoogleSignIn(&GoogleIdentityPayload{
		Subject:       "google-sub-1",
		Email:         "existing@example.com",
		Name:          "Google Name",
		PictureURL:    "https://img.example.com/profile.png",
		EmailVerified: true,
	}, "")
	if err != nil {
		t.Fatalf("expected google sign in to succeed: %v", err)
	}

	identity, err := identityRepo.GetByProviderUserID("google", "google-sub-1")
	if err != nil {
		t.Fatalf("failed to fetch identity: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity to be created")
	}
	if identity.UserID != resp.User.ID {
		t.Fatalf("expected identity to link to existing user %q, got %q", resp.User.ID, identity.UserID)
	}
}

func TestGoogleLoginCreatesNewUser(t *testing.T) {
	service, identityRepo, userRepo := newAuthServiceForTest(t)

	resp, err := service.completeGoogleSignIn(&GoogleIdentityPayload{
		Subject:       "google-sub-2",
		Email:         "new-user@example.com",
		Name:          "New Google User",
		PictureURL:    "https://img.example.com/new-user.png",
		EmailVerified: true,
	}, "refresh-1")
	if err != nil {
		t.Fatalf("expected google sign in to create a new user: %v", err)
	}

	user, err := userRepo.GetUserByID(resp.User.ID)
	if err != nil {
		t.Fatalf("failed to fetch created user: %v", err)
	}
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.PasswordHash != "" {
		t.Fatalf("expected google-created user to have empty password hash, got %q", user.PasswordHash)
	}

	identity, err := identityRepo.GetByProviderUserID("google", "google-sub-2")
	if err != nil {
		t.Fatalf("failed to fetch identity: %v", err)
	}
	if identity.RefreshToken != "refresh-1" {
		t.Fatalf("expected refresh token to be stored, got %q", identity.RefreshToken)
	}
}

func TestGoogleLoginReactivatesInactiveUser(t *testing.T) {
	service, _, userRepo := newAuthServiceForTest(t)

	resp, err := service.Register(RegisterRequest{
		Name:     "Inactive User",
		Email:    "inactive@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	if err := userRepo.SetUserActive(resp.User.ID, false); err != nil {
		t.Fatalf("failed to deactivate user: %v", err)
	}

	_, err = service.completeGoogleSignIn(&GoogleIdentityPayload{
		Subject:       "google-sub-3",
		Email:         "inactive@example.com",
		Name:          "Inactive User",
		EmailVerified: true,
	}, "")
	if err != nil {
		t.Fatalf("expected google sign in to reactivate user: %v", err)
	}

	user, err := userRepo.GetUserByID(resp.User.ID)
	if err != nil {
		t.Fatalf("failed to fetch reactivated user: %v", err)
	}
	if !user.IsActive {
		t.Fatal("expected user to be reactivated")
	}
}

func TestGoogleLoginRejectsUnverifiedEmail(t *testing.T) {
	service, _, _ := newAuthServiceForTest(t)

	_, err := service.completeGoogleSignIn(&GoogleIdentityPayload{
		Subject:       "google-sub-4",
		Email:         "unverified@example.com",
		Name:          "Unverified",
		EmailVerified: false,
	}, "")
	if !errors.Is(err, ErrInvalidGoogleToken) {
		t.Fatalf("expected ErrInvalidGoogleToken, got %v", err)
	}
}

func TestGoogleLoginWithCodeRejectsInvalidState(t *testing.T) {
	service, _, _ := newAuthServiceForTest(t)

	_, err := service.GoogleLoginWithCode(context.Background(), "missing-state", "code")
	if !errors.Is(err, ErrInvalidOAuthState) {
		t.Fatalf("expected invalid oauth state error, got %v", err)
	}
}
