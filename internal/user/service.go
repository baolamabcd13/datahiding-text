package user

import (
	"github.com/baolamabcd13/datahiding-text-app/internal/models"
)

// Service - Interface cho user service
type Service interface {
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
}

// UserService - Triển khai Service interface
type UserService struct {
	repo Repository
}

// NewUserService - Tạo service mới
func NewUserService(repo Repository) Service {
	return &UserService{
		repo: repo,
	}
}

// GetUserByID - Lấy thông tin user theo ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.FindUserByID(id)
}

// GetUserByUsername - Lấy thông tin user theo username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.FindUserByUsername(username)
}

// UpdateUser - Cập nhật thông tin người dùng
func (s *UserService) UpdateUser(user *models.User) error {
	return s.repo.UpdateUser(user)
}