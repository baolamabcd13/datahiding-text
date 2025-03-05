package auth

import (
	"errors"
	"time"

	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"gorm.io/gorm"
)

// Repository - Interface cho auth repository
type Repository interface {
	CreateUser(user *models.User) error
	FindUserByEmail(email string) (*models.User, error)
	FindUserByUsername(username string) (*models.User, error)
	FindUserByCCCD(cccd string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
	CreateVerificationToken(token *models.VerificationToken) error
	FindVerificationToken(token string) (*models.VerificationToken, error)
	DeleteVerificationToken(token string) error
	VerifyUser(userID uint) error
	CreatePasswordResetToken(userID uint, token string, expiresAt time.Time) error
	FindPasswordResetToken(token string) (*models.PasswordResetToken, error)
	DeletePasswordResetToken(token string) error
}

// PostgresRepository - Triển khai Repository interface với PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository - Tạo repository mới
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

// CreateUser - Tạo người dùng mới
func (r *PostgresRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// FindUserByEmail - Tìm người dùng theo email
func (r *PostgresRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByUsername - Tìm người dùng theo username
func (r *PostgresRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByCCCD - Tìm người dùng theo CCCD
func (r *PostgresRepository) FindUserByCCCD(cccd string) (*models.User, error) {
	var user models.User
	result := r.db.Where("cccd = ?", cccd).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByID - Tìm người dùng theo ID
func (r *PostgresRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser - Cập nhật thông tin người dùng
func (r *PostgresRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// CreateVerificationToken - Tạo token xác thực mới
func (r *PostgresRepository) CreateVerificationToken(token *models.VerificationToken) error {
	return r.db.Create(token).Error
}

// FindVerificationToken - Tìm token xác thực
func (r *PostgresRepository) FindVerificationToken(token string) (*models.VerificationToken, error) {
	var verificationToken models.VerificationToken
	result := r.db.Where("token = ?", token).First(&verificationToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &verificationToken, nil
}

// DeleteVerificationToken - Xóa token xác thực
func (r *PostgresRepository) DeleteVerificationToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.VerificationToken{}).Error
}

// VerifyUser - Xác thực user
func (r *PostgresRepository) VerifyUser(userID uint) error {
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Update("email_verified", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// CreatePasswordResetToken - Tạo token đặt lại mật khẩu
func (r *PostgresRepository) CreatePasswordResetToken(userID uint, token string, expiresAt time.Time) error {
	// Xóa token cũ nếu có
	r.db.Where("user_id = ?", userID).Delete(&models.PasswordResetToken{})

	// Tạo token mới
	resetToken := models.PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	result := r.db.Create(&resetToken)
	return result.Error
}

// FindPasswordResetToken - Tìm token đặt lại mật khẩu
func (r *PostgresRepository) FindPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	result := r.db.Where("token = ?", token).First(&resetToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &resetToken, nil
}

// DeletePasswordResetToken - Xóa token đặt lại mật khẩu
func (r *PostgresRepository) DeletePasswordResetToken(token string) error {
	result := r.db.Where("token = ?", token).Delete(&models.PasswordResetToken{})
	return result.Error
} 