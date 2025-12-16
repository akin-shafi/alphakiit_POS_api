package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a POS system user
type User struct {
	ID        uint `gorm:"primaryKey"`
	FirstName string
	LastName  string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	Active    bool
	TenantID  string
	OutletID  *uint
	Role      string // OWNER / MANAGER / CASHIER
	CreatedAt time.Time
	UpdatedAt time.Time
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
