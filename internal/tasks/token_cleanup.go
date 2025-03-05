package tasks

import (
	"log"
	"time"

	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"gorm.io/gorm"
)

// CleanupBlacklistedTokens - Xóa các token đã hết hạn khỏi blacklist
func CleanupBlacklistedTokens(db *gorm.DB) {
	log.Println("Cleaning up expired blacklisted tokens...")
	
	result := db.Where("expires_at < ?", time.Now()).Delete(&models.BlacklistedToken{})
	if result.Error != nil {
		log.Printf("Error cleaning up blacklisted tokens: %v\n", result.Error)
		return
	}
	
	log.Printf("Cleaned up %d expired blacklisted tokens\n", result.RowsAffected)
}

// CleanupPasswordResetTokens - Xóa các token đặt lại mật khẩu đã hết hạn
func CleanupPasswordResetTokens(db *gorm.DB) {
	log.Println("Cleaning up expired password reset tokens...")
	
	result := db.Where("expires_at < ?", time.Now()).Delete(&models.PasswordResetToken{})
	if result.Error != nil {
		log.Printf("Error cleaning up password reset tokens: %v\n", result.Error)
		return
	}
	
	log.Printf("Cleaned up %d expired password reset tokens\n", result.RowsAffected)
}

// ScheduleTokenCleanup - Lên lịch xóa token định kỳ
func ScheduleTokenCleanup(db *gorm.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CleanupBlacklistedTokens(db)
			CleanupPasswordResetTokens(db)
		}
	}()
} 