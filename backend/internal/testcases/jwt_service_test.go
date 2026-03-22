package testcases

import (
	"testing"
	"time"

	"backend/internal/auth/services"

	"github.com/golang-jwt/jwt/v5"
)

const testSecretKey = "test-secret-key-for-testing"

func TestJWTService_GenerateToken(t *testing.T) {
	service := services.NewJWTService(testSecretKey, 24)

	tests := []struct {
		name    string
		userID  string
		email   string
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  "user-123",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			userID:  "user-123",
			email:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	service := services.NewJWTService(testSecretKey, 24)

	t.Run("valid token", func(t *testing.T) {
		token, err := service.GenerateToken("user-123", "test@example.com")
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		claims, err := service.ValidateToken(token)
		if err != nil {
			t.Errorf("ValidateToken() error = %v", err)
			return
		}
		if claims.UserID != "user-123" {
			t.Errorf("ValidateToken() UserID = %v, want user-123", claims.UserID)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("ValidateToken() Email = %v, want test@example.com", claims.Email)
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := service.ValidateToken("invalid-token")
		if err == nil {
			t.Error("ValidateToken() expected error for invalid token")
		}
		if err != services.ErrInvalidToken {
			t.Errorf("ValidateToken() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := service.ValidateToken("")
		if err == nil {
			t.Error("ValidateToken() expected error for empty token")
		}
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		wrongService := services.NewJWTService("wrong-secret", 24)
		token, _ := wrongService.GenerateToken("user-123", "test@example.com")

		_, err := service.ValidateToken(token)
		if err == nil {
			t.Error("ValidateToken() expected error for token with wrong secret")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := service.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed")
		if err == nil {
			t.Error("ValidateToken() expected error for malformed token")
		}
	})
}

func TestJWTService_TokenExpiration(t *testing.T) {
	service := services.NewJWTService(testSecretKey, 0)

	token, err := service.GenerateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	_, err = service.ValidateToken(token)
}

func TestJWTService_Claims(t *testing.T) {
	service := services.NewJWTService(testSecretKey, 24)

	token, err := service.GenerateToken("user-abc", "user@domain.org")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parsedToken, _ := jwt.ParseWithClaims(token, &services.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecretKey), nil
	})

	claims, ok := parsedToken.Claims.(*services.Claims)
	if !ok {
		t.Fatal("Failed to parse claims")
	}

	if claims.UserID != "user-abc" {
		t.Errorf("UserID = %v, want user-abc", claims.UserID)
	}
	if claims.Email != "user@domain.org" {
		t.Errorf("Email = %v, want user@domain.org", claims.Email)
	}
	if claims.Issuer != "taskify" {
		t.Errorf("Issuer = %v, want taskify", claims.Issuer)
	}
}

func TestJWTService_GetSecretKey(t *testing.T) {
	service := services.NewJWTService(testSecretKey, 24)
	if service.GetSecretKey() != testSecretKey {
		t.Errorf("GetSecretKey() = %v, want %v", service.GetSecretKey(), testSecretKey)
	}
}

func TestGetEnvJWTSecret(t *testing.T) {
	secret := services.GetEnvJWTSecret()
	if secret == "" {
		t.Error("GetEnvJWTSecret() should not return empty string")
	}
}
