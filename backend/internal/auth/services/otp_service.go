package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OTPEntry stores an OTP and its metadata
type OTPEntry struct {
	Code       string
	ExpiresAt  time.Time
	ResetToken string // set after OTP is verified
}

// OTPService handles OTP generation, storage, and verification
type OTPService struct {
	store map[string]*OTPEntry // keyed by email
	mu    sync.Mutex
}

// NewOTPService creates a new OTPService
func NewOTPService() *OTPService {
	return &OTPService{
		store: make(map[string]*OTPEntry),
	}
}

// GenerateOTP creates a random 6-digit OTP for the given email (10-minute expiry)
func (s *OTPService) GenerateOTP(email string) (string, error) {
	code, err := generateRandomCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[email] = &OTPEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	return code, nil
}

// VerifyOTP checks the OTP code for an email and returns a reset token if valid
func (s *OTPService) VerifyOTP(email, code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[email]
	if !exists {
		return "", fmt.Errorf("no OTP found for this email")
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(s.store, email)
		return "", fmt.Errorf("OTP has expired")
	}

	if entry.Code != code {
		return "", fmt.Errorf("invalid OTP code")
	}

	// OTP is valid — generate a one-time reset token
	resetToken := uuid.New().String()
	entry.ResetToken = resetToken
	entry.ExpiresAt = time.Now().Add(10 * time.Minute) // extend for password reset

	return resetToken, nil
}

// ValidateResetToken checks if a reset token is valid and returns the associated email
func (s *OTPService) ValidateResetToken(token string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for email, entry := range s.store {
		if entry.ResetToken == token {
			if time.Now().After(entry.ExpiresAt) {
				delete(s.store, email)
				return "", fmt.Errorf("reset token has expired")
			}
			// Consume the token (one-time use)
			delete(s.store, email)
			return email, nil
		}
	}

	return "", fmt.Errorf("invalid reset token")
}

// generateRandomCode generates a cryptographically random numeric code of the given length
func generateRandomCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
