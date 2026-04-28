package helpers

import "strings"

// Validation utilities for input sanitization

// Input size limits
const (
	MaxNameLength     = 100
	MaxTitleLength    = 255
	MaxDescriptionLen = 1000
)

// TrimAndValidate trims whitespace and checks if empty
func TrimAndValidate(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}

// ValidateMaxLength checks if string exceeds max length
func ValidateMaxLength(s string, max int) bool {
	return len(s) <= max
}

// SanitizeInput trims and limits length
func SanitizeInput(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// ValidateName checks name length (max 100)
func ValidateName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > MaxNameLength {
		return "", false
	}
	return name, true
}

// ValidateTitle checks title length (max 255)
func ValidateTitle(title string) (string, bool) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > MaxTitleLength {
		return "", false
	}
	return title, true
}
