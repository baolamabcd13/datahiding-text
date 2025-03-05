package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/baolamabcd13/datahiding-text-app/internal/email"
	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"github.com/baolamabcd13/datahiding-text-app/internal/utils"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Service - Interface cho auth service
type Service interface {
	Register(user *models.User) error
	Login(username, password string) (*models.User, string, error)
	GenerateToken(userID uint) (string, error)
	VerifyToken(tokenString string) (uint, error)
	GenerateVerificationToken(userID uint) (string, error)
	SendVerificationEmail(user *models.User) error
	VerifyEmail(token string) error
	Logout(token string) error
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
}

// Config - Cấu hình cho auth service
type Config struct {
	JWTSecret           string
	JWTExpirationHours  int
	EmailVerificationRequired bool
	AppURL                string
}

// AuthService - Triển khai Service interface
type AuthService struct {
	repo         Repository
	jwtSecret    string
	emailService email.Service
	config       Config
	tokenRepo    TokenRepository
}

// NewAuthService - Tạo service mới
func NewAuthService(repo Repository, jwtSecret string, emailService email.Service, config Config, tokenRepo TokenRepository) Service {
	return &AuthService{
		repo:         repo,
		jwtSecret:    jwtSecret,
		emailService: emailService,
		config:       config,
		tokenRepo:    tokenRepo,
	}
}

// Register - Đăng ký tài khoản mới
func (s *AuthService) Register(user *models.User) error {
	// Kiểm tra username đã tồn tại chưa
	existingUser, err := s.repo.FindUserByUsername(user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// Kiểm tra email đã tồn tại chưa
	existingUser, err = s.repo.FindUserByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already exists")
	}

	// Kiểm tra CCCD đã tồn tại chưa
	if user.CCCD != "" {
		existingUser, err = s.repo.FindUserByCCCD(user.CCCD)
		if err != nil {
			return err
		}
		if existingUser != nil {
			return errors.New("cccd already exists")
		}
	}

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Tạo user
	err = s.repo.CreateUser(user)
	if err != nil {
		return err
	}

	// Nếu yêu cầu xác thực email
	if s.config.EmailVerificationRequired {
		// Tạo token xác thực
		token := utils.GenerateRandomString(32)
		expiresAt := time.Now().Add(24 * time.Hour)

		// Lưu token vào database
		verificationToken := &models.VerificationToken{
			UserID:    user.ID,
			Token:     token,
			ExpiresAt: expiresAt,
		}
		err = s.repo.CreateVerificationToken(verificationToken)
		if err != nil {
			// Log lỗi nhưng không trả về lỗi cho người dùng
			log.Printf("Failed to create verification token: %v", err)
			return nil
		}

		// Gửi email xác thực
		err = s.emailService.SendVerificationEmail(user.Email, user.Name, token)
		if err != nil {
			// Log lỗi nhưng không trả về lỗi cho người dùng
			log.Printf("Failed to send verification email: %v", err)
			return nil
		}
	}

	return nil
}

// Login - Đăng nhập
func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	// Tìm user theo username
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Kiểm tra mật khẩu
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Kiểm tra xác thực email nếu yêu cầu
	if s.config.EmailVerificationRequired && !user.EmailVerified {
		return nil, "", errors.New("email not verified")
	}

	// Tạo JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GenerateToken - Tạo JWT token
func (s *AuthService) GenerateToken(userID uint) (string, error) {
	// Tạo claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Token hết hạn sau 7 ngày
	}

	// Tạo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ký token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken - Xác thực JWT token
func (s *AuthService) VerifyToken(tokenString string) (uint, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra thuật toán ký
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	// Kiểm tra token hợp lệ
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Lấy user_id từ claims
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("invalid token")
}

// GenerateVerificationToken - Tạo token xác thực
func (s *AuthService) GenerateVerificationToken(userID uint) (string, error) {
	// Tạo token ngẫu nhiên
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	
	// Tạo verification token
	verificationToken := &models.VerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Hết hạn sau 24 giờ
	}
	
	// Lưu token vào database
	if err := s.repo.CreateVerificationToken(verificationToken); err != nil {
		return "", err
	}
	
	return token, nil
}

// SendVerificationEmail - Gửi email xác thực
func (s *AuthService) SendVerificationEmail(user *models.User) error {
	// Tạo token xác thực
	token, err := s.GenerateVerificationToken(user.ID)
	if err != nil {
		return err
	}
	
	// Gửi email
	return s.emailService.SendVerificationEmail(user.Email, user.Name, token)
}

// VerifyEmail - Xác thực email
func (s *AuthService) VerifyEmail(token string) error {
	// Tìm token trong database
	verificationToken, err := s.repo.FindVerificationToken(token)
	if err != nil {
		return err
	}
	if verificationToken == nil {
		return errors.New("invalid or expired token")
	}

	// Kiểm tra token hết hạn
	if time.Now().After(verificationToken.ExpiresAt) {
		return errors.New("token has expired")
	}

	// Xác thực user
	err = s.repo.VerifyUser(verificationToken.UserID)
	if err != nil {
		return err
	}

	// Xóa token
	err = s.repo.DeleteVerificationToken(token)
	if err != nil {
		return err
	}

	return nil
}

// Logout - Đăng xuất người dùng bằng cách vô hiệu hóa token
func (s *AuthService) Logout(tokenString string) error {
	// Parse token để lấy thông tin
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra thuật toán ký
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return errors.New("invalid token")
	}

	// Kiểm tra token hợp lệ
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Lấy thông tin từ claims
		userID := uint(claims["user_id"].(float64))
		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		
		// Thêm token vào blacklist
		return s.tokenRepo.AddToBlacklist(tokenString, userID, exp)
	}

	return errors.New("invalid token")
}

// ForgotPassword - Xử lý yêu cầu quên mật khẩu
func (s *AuthService) ForgotPassword(email string) error {
	// Tìm user theo email
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Tạo token ngẫu nhiên
	token := utils.GenerateRandomString(32)

	// Tạo thời gian hết hạn (24 giờ)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Lưu token vào database
	err = s.repo.CreatePasswordResetToken(user.ID, token, expiresAt)
	if err != nil {
		return err
	}

	// Tạo link đặt lại mật khẩu
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.AppURL, token)

	// Gửi email
	err = s.emailService.SendPasswordResetEmail(user.Email, user.Name, resetLink)
	if err != nil {
		return err
	}

	return nil
}

// ResetPassword - Đặt lại mật khẩu với token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Tìm token trong database
	resetToken, err := s.repo.FindPasswordResetToken(token)
	if err != nil {
		return err
	}
	if resetToken == nil {
		return errors.New("invalid or expired token")
	}

	// Kiểm tra token hết hạn
	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("token has expired")
	}

	// Tìm user
	user, err := s.repo.FindUserByID(resetToken.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Hash mật khẩu mới
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Cập nhật mật khẩu
	user.Password = string(hashedPassword)
	err = s.repo.UpdateUser(user)
	if err != nil {
		return err
	}

	// Xóa token
	err = s.repo.DeletePasswordResetToken(token)
	if err != nil {
		return err
	}

	return nil
}

// EmailService - Interface cho email service
type EmailService interface {
	SendEmail(to, subject, body string) error
	SendVerificationEmail(to, name, verificationLink string) error
	SendPasswordResetEmail(to, name, resetLink string) error
}

// generateJWT - Tạo JWT token
func (s *AuthService) generateJWT(user *models.User) (string, error) {
	// Tạo claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.config.JWTExpirationHours)).Unix(),
	}

	// Tạo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ký token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
} 