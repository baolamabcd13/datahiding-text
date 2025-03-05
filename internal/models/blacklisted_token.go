package models

import (
	"time"
)

// BlacklistedToken - Model cho token bị vô hiệu hóa
type BlacklistedToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"type:text;not null;uniqueIndex" json:"token"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
} 