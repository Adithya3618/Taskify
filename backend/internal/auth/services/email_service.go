package services

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailService handles sending emails via SMTP
type EmailService struct {
	host     string
	port     string
	email    string
	password string
}

// NewEmailService creates a new EmailService with SMTP configuration
func NewEmailService() *EmailService {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	return &EmailService{
		host:     host,
		port:     port,
		email:    os.Getenv("SMTP_EMAIL"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

// SendOTP sends a verification OTP code to the specified email address
func (s *EmailService) SendOTP(toEmail, otp string) error {
	auth := smtp.PlainAuth("", s.email, s.password, s.host)

	subject := "Taskify - Password Reset Code"
	body := fmt.Sprintf(
		"Hello,\n\nYour password reset verification code is: %s\n\nThis code will expire in 10 minutes.\n\nIf you did not request this, please ignore this email.\n\n- Taskify Team",
		otp,
	)

	msg := fmt.Sprintf(
		"From: Taskify <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		s.email, toEmail, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, s.email, []string{toEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
