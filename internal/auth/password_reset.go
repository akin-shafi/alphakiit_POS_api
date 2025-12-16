// auth/password_reset.go
package auth

import (
	"crypto/rand"
	"log"
	"math/big"
	"pos-fiber-app/internal/email"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"pos-fiber-app/internal/user" // assuming UserService or direct access; adjust if needed
)

const otpLength = 6
const otpExpiryMinutes = 10

// PasswordResetOTP model
type PasswordResetOTP struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"index;size:255"`
	OTPHash   string `gorm:"size:255"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Migrate password reset table
func MigratePasswordReset(db *gorm.DB) error {
	return db.AutoMigrate(&PasswordResetOTP{})
}

// generateOTP creates a numeric 6-digit OTP
func generateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, otpLength)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}


// ForgotPasswordHandler godoc
// @Summary Request password reset OTP
// @Description Send a 6-digit OTP to the user's email for password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string} true "Email address"
// @Success 200 {object} map[string]string "message: OTP sent"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/forgot-password [post]
func ForgotPasswordHandler(db *gorm.DB) fiber.Handler {
	sender := email.NewSender(email.LoadConfig()) // or inject via dependency

	return func(c *fiber.Ctx) error {
		var payload struct {
			Email string `json:"email" validate:"required,email"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		var u user.User
		if err := db.Where("email = ? AND active = true", payload.Email).First(&u).Error; err != nil {
			// Do NOT reveal if email exists
			return c.JSON(fiber.Map{"message": "If the email exists, an OTP has been sent"})
		}

		otp, err := generateOTP()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate OTP"})
		}

		hashed, err := user.HashPassword(otp)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to process OTP"})
		}

		// Delete any old OTP for this email
		db.Where("email = ?", payload.Email).Delete(&PasswordResetOTP{})

		reset := PasswordResetOTP{
			Email:     payload.Email,
			OTPHash:   hashed,
			ExpiresAt: time.Now().Add(otpExpiryMinutes * time.Minute),
		}
		db.Create(&reset)

		// TODO: Replace with real email service (SMTP, SendGrid, SES, etc.)
		// log.Printf("=== PASSWORD RESET OTP ===\nEmail: %s\nOTP: %s\nExpires in %d minutes\n", payload.Email, otp, otpExpiryMinutes)
		if err := sender.SendPasswordResetOTP(payload.Email, otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
			// Still return success to avoid info disclosure
		}
		return c.JSON(fiber.Map{"message": "If the email exists, an OTP has been sent"})
	}
}


// VerifyOTPHandler godoc
// @Summary Verify OTP
// @Description Verify the OTP sent to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string,otp=string} true "Email and OTP"
// @Success 200 {object} map[string]string "message: OTP verified"
// @Failure 400 {object} map[string]string "error: Invalid or expired OTP"
// @Router /auth/verify-otp [post]
func VerifyOTPHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			Email string `json:"email" validate:"required,email"`
			OTP   string `json:"otp" validate:"required,len=6"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		var reset PasswordResetOTP
		err := db.Where("email = ? AND expires_at > ?", payload.Email, time.Now()).
			Order("created_at desc").
			First(&reset).Error

		if err != nil || !user.CheckPassword(reset.OTPHash, payload.OTP) {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired OTP"})
		}

		// OTP valid → optionally delete it after verification (one-time use)
		db.Delete(&reset)

		return c.JSON(fiber.Map{"message": "OTP verified successfully"})
	}
}


// ResetPasswordHandler godoc
// @Summary Reset password with OTP
// @Description Set a new password using the valid OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string,otp=string,new_password=string} true "Reset payload"
// @Success 200 {object} map[string]string "message: Password reset successfully"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/reset-password [post]

func ResetPasswordHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			Email       string `json:"email" validate:"required,email"`
			OTP         string `json:"otp" validate:"required,len=6"`
			NewPassword string `json:"new_password" validate:"required,min=8"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		var reset PasswordResetOTP
		err := db.Where("email = ? AND expires_at > ?", payload.Email, time.Now()).
			Order("created_at desc").
			First(&reset).Error

		if err != nil || !user.CheckPassword(reset.OTPHash, payload.OTP) {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired OTP"})
		}

		var u user.User
		if err := db.Where("email = ?", payload.Email).First(&u).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		hashed, err := user.HashPassword(payload.NewPassword)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
		}

		u.Password = hashed
		db.Save(&u)

		// Clean up OTP
		db.Delete(&reset)

		return c.JSON(fiber.Map{"message": "Password reset successfully"})
	}
}
