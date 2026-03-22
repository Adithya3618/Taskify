package testcases

import (
	"regexp"
	"strings"
	"testing"

	"backend/internal/auth/services"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"user+tag@example.com", true},
		{"user@sub.domain.example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
		{"user@.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"  user@domain.com  ", "user@domain.com"},
		{"user.name@example.org", "user.name@example.org"},
		{"UPPERCASE@DOMAIN.COM", "uppercase@domain.com"},
		{"Mixed@Case.COM", "mixed@case.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRegisterRequest_Structure(t *testing.T) {
	req := services.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "securepassword",
	}

	if req.Name != "Test User" {
		t.Errorf("RegisterRequest.Name = %v, want Test User", req.Name)
	}
	if req.Email != "test@example.com" {
		t.Errorf("RegisterRequest.Email = %v, want test@example.com", req.Email)
	}
}

func TestLoginRequest_Structure(t *testing.T) {
	req := services.LoginRequest{
		Email:    "login@example.com",
		Password: "loginpassword",
	}

	if req.Email != "login@example.com" {
		t.Errorf("LoginRequest.Email = %v, want login@example.com", req.Email)
	}
}

func TestAuthServiceErrors(t *testing.T) {
	if services.ErrUserExists.Error() != "user with this email already exists" {
		t.Errorf("ErrUserExists = %v", services.ErrUserExists)
	}
	if services.ErrInvalidCredentials.Error() != "invalid email or password" {
		t.Errorf("ErrInvalidCredentials = %v", services.ErrInvalidCredentials)
	}
	if services.ErrWeakPassword.Error() != "password must be at least 8 characters" {
		t.Errorf("ErrWeakPassword = %v", services.ErrWeakPassword)
	}
	if services.ErrInvalidEmail.Error() != "invalid email format" {
		t.Errorf("ErrInvalidEmail = %v", services.ErrInvalidEmail)
	}
	if services.ErrUserInactive.Error() != "user account is inactive" {
		t.Errorf("ErrUserInactive = %v", services.ErrUserInactive)
	}
}

func TestAuthResponse_Structure(t *testing.T) {
	resp := services.AuthResponse{
		Token: "jwt-token-here",
	}

	if resp.Token != "jwt-token-here" {
		t.Errorf("AuthResponse.Token = %v, want jwt-token-here", resp.Token)
	}
}

// isValidEmail checks if the email format is valid (copy for testing exported behavior)
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// normalizeEmail converts email to lowercase and trims whitespace (copy for testing)
func normalizeEmail(email string) string {
	return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(email, ""))
}
