package testcases

import (
	"database/sql"
	"os"
	"testing"

	"backend/internal/auth/models"
	"backend/internal/auth/repository"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpFile, err := os.CreateTemp("", "test_user_repo_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("sql.Open() error = %v", err)
	}

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	if repo == nil {
		t.Error("NewUserRepository() should not return nil")
	}
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		ID:           "test-user-1",
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
		IsActive:     true,
	}

	err := repo.CreateUser(user)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
	}

	// Verify user was created
	retrieved, err := repo.GetUserByEmail("test@example.com")
	if err != nil {
		t.Errorf("GetUserByEmail() error = %v", err)
	}
	if retrieved == nil {
		t.Error("GetUserByEmail() should return the created user")
	}
	if retrieved.ID != user.ID {
		t.Errorf("User ID = %v, want %v", retrieved.ID, user.ID)
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("returns nil for non-existent email", func(t *testing.T) {
		user, err := repo.GetUserByEmail("nonexistent@example.com")
		if err != nil {
			t.Errorf("GetUserByEmail() error = %v", err)
		}
		if user != nil {
			t.Error("GetUserByEmail() should return nil for non-existent email")
		}
	})

	t.Run("returns user for existing email", func(t *testing.T) {
		user := &models.User{
			ID:           "user-email-test",
			Name:         "Email Test",
			Email:        "email@test.com",
			PasswordHash: "hash",
			Role:         models.RoleUser,
			IsActive:     true,
		}
		repo.CreateUser(user)

		retrieved, err := repo.GetUserByEmail("email@test.com")
		if err != nil {
			t.Errorf("GetUserByEmail() error = %v", err)
		}
		if retrieved == nil {
			t.Error("GetUserByEmail() should return the user")
		}
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("returns nil for non-existent ID", func(t *testing.T) {
		user, err := repo.GetUserByID("non-existent-id")
		if err != nil {
			t.Errorf("GetUserByID() error = %v", err)
		}
		if user != nil {
			t.Error("GetUserByID() should return nil for non-existent ID")
		}
	})

	t.Run("returns user for existing ID", func(t *testing.T) {
		user := &models.User{
			ID:           "user-id-test-123",
			Name:         "ID Test",
			Email:        "idtest@example.com",
			PasswordHash: "hash",
			Role:         models.RoleUser,
			IsActive:     true,
		}
		repo.CreateUser(user)

		retrieved, err := repo.GetUserByID("user-id-test-123")
		if err != nil {
			t.Errorf("GetUserByID() error = %v", err)
		}
		if retrieved == nil {
			t.Error("GetUserByID() should return the user")
		}
	})
}

func TestUserRepository_EmailExists(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("returns false for non-existent email", func(t *testing.T) {
		exists, err := repo.EmailExists("nonexistent@example.com")
		if err != nil {
			t.Errorf("EmailExists() error = %v", err)
		}
		if exists {
			t.Error("EmailExists() should return false for non-existent email")
		}
	})

	t.Run("returns true for existing email", func(t *testing.T) {
		user := &models.User{
			ID:           "user-exists-test",
			Name:         "Exists Test",
			Email:        "exists@example.com",
			PasswordHash: "hash",
			Role:         models.RoleUser,
			IsActive:     true,
		}
		repo.CreateUser(user)

		exists, err := repo.EmailExists("exists@example.com")
		if err != nil {
			t.Errorf("EmailExists() error = %v", err)
		}
		if !exists {
			t.Error("EmailExists() should return true for existing email")
		}
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		ID:           "user-update-pwd",
		Name:         "Update Password Test",
		Email:        "updatepwd@example.com",
		PasswordHash: "oldhash",
		Role:         models.RoleUser,
		IsActive:     true,
	}
	repo.CreateUser(user)

	t.Run("updates password successfully", func(t *testing.T) {
		err := repo.UpdatePassword("updatepwd@example.com", "newhash")
		if err != nil {
			t.Errorf("UpdatePassword() error = %v", err)
		}

		retrieved, _ := repo.GetUserByEmail("updatepwd@example.com")
		if retrieved.PasswordHash != "newhash" {
			t.Errorf("PasswordHash = %v, want newhash", retrieved.PasswordHash)
		}
	})

	t.Run("fails for non-existent email", func(t *testing.T) {
		err := repo.UpdatePassword("nonexistent@example.com", "hash")
		if err == nil {
			t.Error("UpdatePassword() should fail for non-existent email")
		}
	})
}
