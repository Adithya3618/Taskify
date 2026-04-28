package helpers

import "strings"

// Validation utilities for input sanitization

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
