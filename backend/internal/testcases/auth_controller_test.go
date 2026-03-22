package testcases

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/auth/controller"
)

func TestAuthController_Register_InvalidJSON(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		ctrl.Register(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Register() status = %v, want 400", w.Code)
		}
	})
}

func TestAuthController_Login_InvalidJSON(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		ctrl.Login(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Login() status = %v, want 400", w.Code)
		}
	})
}

func TestAuthController_GetMe_NoUserID(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 401 when no user ID in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me", nil)

		ctrl.GetMe(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GetMe() status = %v, want 401", w.Code)
		}
	})
}

func TestNewAuthController(t *testing.T) {
	ctrl := controller.NewAuthController(nil)
	if ctrl == nil {
		t.Error("NewAuthController() should not return nil")
	}
}

func TestAuthController_ForgotPassword(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")

		ctrl.ForgotPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("ForgotPassword() status = %v, want 400", w.Code)
		}
	})

	t.Run("returns 400 for empty email", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]string{"email": ""}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		ctrl.ForgotPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("ForgotPassword() status = %v, want 400", w.Code)
		}
	})
}

func TestAuthController_VerifyOTP(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/verify-otp", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")

		ctrl.VerifyOTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("VerifyOTP() status = %v, want 400", w.Code)
		}
	})

	t.Run("returns 400 for missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]string{"email": "", "code": ""}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/verify-otp", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		ctrl.VerifyOTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("VerifyOTP() status = %v, want 400", w.Code)
		}
	})
}

func TestAuthController_ResetPassword(t *testing.T) {
	ctrl := &controller.AuthController{}

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/reset-password", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")

		ctrl.ResetPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("ResetPassword() status = %v, want 400", w.Code)
		}
	})

	t.Run("returns 400 for missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]string{"reset_token": "", "new_password": ""}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/reset-password", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		ctrl.ResetPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("ResetPassword() status = %v, want 400", w.Code)
		}
	})
}
