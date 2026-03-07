package services

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"backend/internal/auth/models"
	"backend/internal/auth/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo     *repository.UserRepository
	jwtService   *JWTService
	otpService   *OTPService
	emailService *EmailService
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo *repository.UserRepository, jwtService *JWTService, otpService *OTPService, emailService *EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtService:   jwtService,
		otpService:   otpService,
		emailService: emailService,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  models.UserResponse `json:"user"`
	Token string              `json:"token"`
}

// Register creates a new user account
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateRegisterInput(req); err != nil {
		return nil, err
	}

	// Normalize email
	email := normalizeEmail(req.Email)

	// Check if user already exists
	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %v", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	now := time.Now()
	user := &models.User{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleUser, // Default role
		IsActive:     true,            // Active by default
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to database
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	// Normalize email
	email := normalizeEmail(req.Email)

	// Find user by email
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// ForgotPassword generates an OTP and sends it to the user's email
func (s *AuthService) ForgotPassword(email string) error {
	email = normalizeEmail(email)

	// Check if user exists
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to check user: %v", err)
	}
	if user == nil {
		// Return nil to avoid email enumeration
		return nil
	}

	// Generate OTP
	otp, err := s.otpService.GenerateOTP(email)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %v", err)
	}

	// Send OTP via email
	if err := s.emailService.SendOTP(email, otp); err != nil {
		return fmt.Errorf("failed to send OTP email: %v", err)
	}

	return nil
}

// VerifyResetOTP verifies the OTP and returns a reset token
func (s *AuthService) VerifyResetOTP(email, code string) (string, error) {
	email = normalizeEmail(email)

	resetToken, err := s.otpService.VerifyOTP(email, code)
	if err != nil {
		return "", err
	}

	return resetToken, nil
}

// ResetPassword resets the user's password using a valid reset token
func (s *AuthService) ResetPassword(resetToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	// Validate reset token and get email
	email, err := s.otpService.ValidateResetToken(resetToken)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password in database
	if err := s.userRepo.UpdatePassword(email, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

// validateRegisterInput validates the registration input
func (s *AuthService) validateRegisterInput(req RegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Email == "" {
		return ErrInvalidEmail
	}
	if !isValidEmail(req.Email) {
		return ErrInvalidEmail
	}
	if len(req.Password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// normalizeEmail converts email to lowercase and trims whitespace
func normalizeEmail(email string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(email, "")
}
