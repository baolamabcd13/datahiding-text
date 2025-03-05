package models

import (
	"gorm.io/gorm"
)

// User - Model cho user
type User struct {
	gorm.Model
	Username      string `gorm:"uniqueIndex;not null"`
	Password      string `gorm:"not null"`
	Email         string `gorm:"uniqueIndex;not null"`
	Name          string
	Phone         string
	CCCD          string `gorm:"uniqueIndex"`
	EmailVerified bool   `gorm:"default:false"`
	Avatar        string
} 