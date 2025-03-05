package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

// Config - Cấu hình cho email service
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	AppURL   string
}

// Service - Interface cho email service
type Service interface {
	SendVerificationEmail(to, name, token string) error
	SendPasswordResetEmail(to, name, resetToken string) error
}

// EmailService - Triển khai Service interface
type EmailService struct {
	config Config
}

// NewEmailService - Tạo service mới
func NewEmailService(config Config) Service {
	return &EmailService{
		config: config,
	}
}

// SendVerificationEmail - Gửi email xác thực
func (s *EmailService) SendVerificationEmail(to, name, token string) error {
	// Kiểm tra email hợp lệ
	if to == "" {
		return errors.New("recipient email address is empty")
	}

	// Kiểm tra email người gửi
	if s.config.From == "" {
		return errors.New("sender email address is empty")
	}

	// Tạo nội dung email
	subject := "Xác thực tài khoản Dating Text App"
	
	// Tạo URL xác thực
	verificationURL := fmt.Sprintf("%s/api/auth/verify-email?token=%s", s.config.AppURL, token)
	
	// In ra log để debug
	fmt.Printf("Verification Email URL: %s\n", verificationURL)
	fmt.Printf("Sending email to: %s, From: %s\n", to, s.config.From)
	
	// Đường dẫn đến template
	templatePath := filepath.Join("templates", "verification_email.html")
	fmt.Printf("Template path: %s\n", templatePath)
	
	// Parse template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}
	
	// Render template
	var body bytes.Buffer
	err = tmpl.Execute(&body, struct {
		Name            string
		VerificationURL string
	}{
		Name:            name,
		VerificationURL: verificationURL,
	})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}
	
	// Tạo message
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())
	
	// Tạo dialer
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	
	// Gửi email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}

// SendPasswordResetEmail - Gửi email đặt lại mật khẩu
func (s *EmailService) SendPasswordResetEmail(to, name, resetToken string) error {
	// Kiểm tra email hợp lệ
	if to == "" {
		return errors.New("recipient email address is empty")
	}

	// Kiểm tra email người gửi
	if s.config.From == "" {
		return errors.New("sender email address is empty")
	}

	// Tạo nội dung email
	subject := "Đặt lại mật khẩu Dating Text App"
	
	// Tạo URL đặt lại mật khẩu
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.AppURL, resetToken)
	
	// In ra log để debug
	fmt.Printf("Reset Password URL: %s\n", resetURL)
	fmt.Printf("Sending email to: %s, From: %s\n", to, s.config.From)
	
	// Đường dẫn đến template
	templatePath := filepath.Join("templates", "reset_password.html")
	fmt.Printf("Template path: %s\n", templatePath)
	
	// Parse template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}
	
	// Render template
	var body bytes.Buffer
	err = tmpl.Execute(&body, struct {
		Name     string
		ResetURL string
	}{
		Name:     name,
		ResetURL: resetURL,
	})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}
	
	// Tạo message
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())
	
	// Tạo dialer
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	
	// Gửi email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
} 