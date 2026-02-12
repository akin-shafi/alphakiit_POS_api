package auth

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token stored in DB
type RefreshToken struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Token      string    `gorm:"uniqueIndex;size:512" json:"token"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AutoMigrate helper
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&RefreshToken{})
}
