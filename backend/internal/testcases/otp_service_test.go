package testcases

import (
	"strings"
	"testing"

	"backend/internal/auth/services"
)

func TestOTPService_GenerateOTP(t *testing.T) {
	service := services.NewOTPService()

	t.Run("generates 6-digit OTP", func(t *testing.T) {
		otp, err := service.GenerateOTP("test@example.com")
		if err != nil {
			t.Errorf("GenerateOTP() error = %v", err)
			return
		}
		if len(otp) != 6 {
			t.Errorf("GenerateOTP() length = %v, want 6", len(otp))
		}
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("GenerateOTP() contains non-digit: %c", c)
			}
		}
	})

	t.Run("generates unique OTPs", func(t *testing.T) {
		otp1, _ := service.GenerateOTP("test@example.com")
		otp2, _ := service.GenerateOTP("test@example.com")
		if otp1 == otp2 {
			t.Error("GenerateOTP() should generate unique OTPs for same email")
		}
	})
}

func TestOTPService_VerifyOTP(t *testing.T) {
	service := services.NewOTPService()

	t.Run("valid OTP verification", func(t *testing.T) {
		email := "verify@example.com"
		otp, _ := service.GenerateOTP(email)

		resetToken, err := service.VerifyOTP(email, otp)
		if err != nil {
			t.Errorf("VerifyOTP() error = %v", err)
			return
		}
		if resetToken == "" {
			t.Error("VerifyOTP() should return a reset token")
		}
	})

	t.Run("invalid OTP code", func(t *testing.T) {
		email := "invalid@example.com"
		service.GenerateOTP(email)

		_, err := service.VerifyOTP(email, "000000")
		if err == nil {
			t.Error("VerifyOTP() should return error for invalid OTP")
		}
	})

	t.Run("non-existent email", func(t *testing.T) {
		_, err := service.VerifyOTP("nonexistent@example.com", "123456")
		if err == nil {
			t.Error("VerifyOTP() should return error for non-existent email")
		}
	})
}

func TestOTPService_ValidateResetToken(t *testing.T) {
	service := services.NewOTPService()

	t.Run("valid reset token", func(t *testing.T) {
		email := "reset@example.com"
		otp, _ := service.GenerateOTP(email)
		resetToken, _ := service.VerifyOTP(email, otp)

		retrievedEmail, err := service.ValidateResetToken(resetToken)
		if err != nil {
			t.Errorf("ValidateResetToken() error = %v", err)
			return
		}
		if retrievedEmail != email {
			t.Errorf("ValidateResetToken() email = %v, want %v", retrievedEmail, email)
		}
	})

	t.Run("invalid reset token", func(t *testing.T) {
		_, err := service.ValidateResetToken("invalid-reset-token")
		if err == nil {
			t.Error("ValidateResetToken() should return error for invalid token")
		}
		if !strings.Contains(err.Error(), "invalid reset token") {
			t.Errorf("ValidateResetToken() error message should contain 'invalid reset token'")
		}
	})

	t.Run("one-time use token", func(t *testing.T) {
		email := "onetime@example.com"
		otp, _ := service.GenerateOTP(email)
		resetToken, _ := service.VerifyOTP(email, otp)

		_, err := service.ValidateResetToken(resetToken)
		if err != nil {
			t.Errorf("ValidateResetToken() first use error = %v", err)
		}

		_, err = service.ValidateResetToken(resetToken)
		if err == nil {
			t.Error("ValidateResetToken() should fail on second use")
		}
	})
}
