package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a POS system user
type User struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Phone             string    `json:"phone"`
	Email             string    `gorm:"uniqueIndex" json:"email"`
	Password          string    `json:"-"`
	Active            bool      `json:"active"`
	TenantID          string    `json:"tenant_id"`
	OutletID          *uint     `json:"outlet_id"`
	Role              string    `json:"role"` // OWNER / MANAGER / CASHIER
	IsVerified        bool      `gorm:"default:false" json:"is_verified"`
	VerificationToken string    `json:"-"` // Temporary token if needed, or just rely on OTP table
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

// HashPassword hashes a plain password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword compares hashed password with plain text
func CheckPassword(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

// AutoMigrate helper
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
