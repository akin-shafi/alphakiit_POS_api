package auth

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token stored in DB
type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"uniqueIndex;size:512"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AutoMigrate helper
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&RefreshToken{})
}
