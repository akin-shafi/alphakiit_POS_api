package notification

import (
	"time"

	"gorm.io/gorm"
)

type DeviceToken struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	BusinessID uint      `gorm:"index" json:"business_id"`
	Token      string    `gorm:"uniqueIndex" json:"token"`
	DeviceType string    `json:"device_type"` // ios, android, web
	LastUsed   time.Time `json:"last_used"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DeviceToken{})
}
