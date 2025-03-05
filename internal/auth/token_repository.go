package auth

import (
	"time"

	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"gorm.io/gorm"
)

// TokenRepository - Interface cho token repository
type TokenRepository interface {
	AddToBlacklist(token string, userID uint, expiresAt time.Time) error
	IsBlacklisted(token string) (bool, error)
}

// PostgresTokenRepository - Triển khai TokenRepository interface với PostgreSQL
type PostgresTokenRepository struct {
	db *gorm.DB
}

// NewPostgresTokenRepository - Tạo token repository mới
func NewPostgresTokenRepository(db *gorm.DB) TokenRepository {
	return &PostgresTokenRepository{
		db: db,
	}
}

// AddToBlacklist - Thêm token vào blacklist
func (r *PostgresTokenRepository) AddToBlacklist(token string, userID uint, expiresAt time.Time) error {
	blacklistedToken := models.BlacklistedToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	
	result := r.db.Create(&blacklistedToken)
	return result.Error
}

// IsBlacklisted - Kiểm tra xem token có trong blacklist không
func (r *PostgresTokenRepository) IsBlacklisted(token string) (bool, error) {
	var count int64
	result := r.db.Model(&models.BlacklistedToken{}).Where("token = ?", token).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
} 