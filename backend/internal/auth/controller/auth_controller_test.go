package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/auth/services"
)

func TestGoogleLoginWithIDTokenRequiresToken(t *testing.T) {
	controller := NewAuthController(&services.AuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/google/id-token", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	controller.GoogleLoginWithIDToken(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestGoogleCallbackRequiresStateAndCode(t *testing.T) {
	controller := NewAuthController(&services.AuthService{})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback", nil)
	recorder := httptest.NewRecorder()

	controller.GoogleCallback(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleAuthErrorForGoogleNotConfigured(t *testing.T) {
	controller := NewAuthController(&services.AuthService{})
	recorder := httptest.NewRecorder()

	controller.GoogleLoginRedirect(recorder, httptest.NewRequest(http.MethodGet, "/api/auth/google/login", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message in response")
	}
}
