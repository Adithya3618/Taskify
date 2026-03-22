package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var ErrInvalidOAuthState = errors.New("invalid oauth state")

type stateEntry struct {
	expiresAt time.Time
}

// OAuthStateService stores short-lived OAuth state values in memory.
type OAuthStateService struct {
	mu     sync.Mutex
	store  map[string]stateEntry
	expiry time.Duration
}

// NewOAuthStateService creates a new OAuthStateService.
func NewOAuthStateService(expiry time.Duration) *OAuthStateService {
	return &OAuthStateService{
		store:  make(map[string]stateEntry),
		expiry: expiry,
	}
}

// Generate creates and stores a new state token.
func (s *OAuthStateService) Generate() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	state := base64.RawURLEncoding.EncodeToString(token)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()
	s.store[state] = stateEntry{expiresAt: time.Now().Add(s.expiry)}
	return state, nil
}

// Consume validates and removes a state token.
func (s *OAuthStateService) Consume(state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()
	entry, exists := s.store[state]
	if !exists || time.Now().After(entry.expiresAt) {
		delete(s.store, state)
		return ErrInvalidOAuthState
	}

	delete(s.store, state)
	return nil
}

func (s *OAuthStateService) cleanupExpiredLocked() {
	now := time.Now()
	for state, entry := range s.store {
		if now.After(entry.expiresAt) {
			delete(s.store, state)
		}
	}
}
