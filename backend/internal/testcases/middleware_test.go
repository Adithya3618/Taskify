package testcases

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/auth/middleware"
	"backend/internal/auth/services"

	"github.com/gorilla/mux"
)

func TestGetUserID(t *testing.T) {
	t.Run("returns user ID from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", "user-123")
		result := middleware.GetUserID(ctx)
		if result != "user-123" {
			t.Errorf("GetUserID() = %v, want user-123", result)
		}
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		result := middleware.GetUserID(ctx)
		if result != "" {
			t.Errorf("GetUserID() = %v, want empty string", result)
		}
	})
}

func TestGetUserEmail(t *testing.T) {
	t.Run("returns user email from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_email", "test@example.com")
		result := middleware.GetUserEmail(ctx)
		if result != "test@example.com" {
			t.Errorf("GetUserEmail() = %v, want test@example.com", result)
		}
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		result := middleware.GetUserEmail(ctx)
		if result != "" {
			t.Errorf("GetUserEmail() = %v, want empty string", result)
		}
	})
}

func TestJWTAuthMiddleware(t *testing.T) {
	jwtService := services.NewJWTService("test-secret", 24)

	t.Run("allows request with valid token", func(t *testing.T) {
		token, _ := jwtService.GenerateToken("user-123", "test@example.com")

		router := mux.NewRouter()
		router.Use(middleware.JWTAuthMiddleware(jwtService))
		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			userID := middleware.GetUserID(r.Context())
			email := middleware.GetUserEmail(r.Context())
			w.Write([]byte(userID + ":" + email))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %v, want 200", w.Code)
		}
	})

	t.Run("rejects request without Authorization header", func(t *testing.T) {
		router := mux.NewRouter()
		router.Use(middleware.JWTAuthMiddleware(jwtService))
		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Status = %v, want 401", w.Code)
		}
	})

	t.Run("rejects request with invalid token", func(t *testing.T) {
		router := mux.NewRouter()
		router.Use(middleware.JWTAuthMiddleware(jwtService))
		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Status = %v, want 401", w.Code)
		}
	})

	t.Run("rejects request with wrong secret", func(t *testing.T) {
		wrongService := services.NewJWTService("wrong-secret", 24)
		token, _ := wrongService.GenerateToken("user-123", "test@example.com")

		router := mux.NewRouter()
		router.Use(middleware.JWTAuthMiddleware(jwtService))
		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Status = %v, want 401", w.Code)
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("writes error response", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeError(w, http.StatusBadRequest, "test error message")

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %v, want 400", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", contentType)
		}

		body := w.Body.String()
		if body == "" {
			t.Error("Response body should not be empty")
		}
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
