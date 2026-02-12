package otp

import (
	"time"

	"gorm.io/gorm"
)

type OTPType string

const (
	TypeVerification  OTPType = "VERIFICATION"
	TypePasswordReset OTPType = "PASSWORD_RESET"
)

type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index;not null"`
	Code      string    `gorm:"not null"`
	Type      OTPType   `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// Migrate - helper to migrate the schema
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&OTP{})
}
