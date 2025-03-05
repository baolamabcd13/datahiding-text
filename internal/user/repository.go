package user

import (
	"errors"

	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"gorm.io/gorm"
)

// Repository - Interface cho user repository
type Repository interface {
	FindUserByID(id uint) (*models.User, error)
	FindUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
}

// PostgresRepository - Triển khai Repository interface với PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository - Tạo repository mới
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{
		db: db,
	}
}

// FindUserByID - Tìm user theo ID
func (r *PostgresRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByUsername - Tìm user theo username
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

// UpdateUser - Cập nhật thông tin người dùng
func (r *PostgresRepository) UpdateUser(user *models.User) error {
	// Sử dụng Updates thay vì Save để chỉ cập nhật các trường được chỉ định
	// Save sẽ cập nhật tất cả các trường, bao gồm cả các trường zero value
	result := r.db.Model(user).Updates(map[string]interface{}{
		"username": user.Username,
		"name":     user.Name,
		"phone":    user.Phone,
		"avatar":   user.Avatar,
	})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("no rows affected, update failed")
	}
	
	return nil
}