package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secretKey      string
	expirationTime time.Duration
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWTService
func NewJWTService(secretKey string, expirationHours int) *JWTService {
	return &JWTService{
		secretKey:      secretKey,
		expirationTime: time.Duration(expirationHours) * time.Hour,
	}
}

// GenerateToken creates a new JWT token for a user
func (s *JWTService) GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "taskify",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetSecretKey returns the secret key (for testing)
func (s *JWTService) GetSecretKey() string {
	return s.secretKey
}

// GetEnvJWTSecret retrieves JWT secret from environment or returns default for development
func GetEnvJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// WARNING: Using a default secret in development only
		// In production, JWT_SECRET must be set
		return "taskify-dev-secret-change-in-production"
	}
	return secret
}
