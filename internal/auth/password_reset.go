package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/user"
)

// ForgotPasswordHandler godoc
// @Summary Request Password Reset
// @Description Generates a password reset OTP and sends it to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string} true "Email payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/forgot-password [post]
func ForgotPasswordHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			Email string `json:"email"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// Check if user exists
		var u user.User
		if err := db.Where("email = ? AND active = true", payload.Email).First(&u).Error; err != nil {
			// Return generic message
			return c.JSON(fiber.Map{"message": "If the email exists, an OTP has been sent"})
		}

		// Generate OTP
		code, err := otp.GenerateOTP()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate OTP"})
		}

		// Save OTP
		otpEntry := otp.OTP{
			Email:     payload.Email,
			Code:      code,
			Type:      otp.TypePasswordReset,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		if err := db.Create(&otpEntry).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save OTP"})
		}

		// Send Email
		sender := email.NewSender(email.LoadConfig())
		body := fmt.Sprintf("<h1>Reset your password</h1><p>Your OTP Code is: <b>%s</b></p><p>It expires in 15 minutes.</p>", code)

		go func() {
			if err := sender.SendCustomEmail(payload.Email, "Reset Your Password", body); err != nil {
				log.Printf("Failed to send reset email: %v", err)
			}
		}()

		return c.JSON(fiber.Map{"message": "If the email exists, an OTP has been sent"})
	}
}

// VerifyResetOTPHandler godoc
// @Summary Verify Reset OTP
// @Description Checks if the password reset OTP is valid
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string,otp=string} true "Verification payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/verify-reset-otp [post]
func VerifyResetOTPHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			Email string `json:"email"`
			OTP   string `json:"otp"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		var otpRecord otp.OTP
		if err := db.Where("email = ? AND code = ? AND type = ? AND used = false AND expires_at > ?",
			payload.Email, payload.OTP, otp.TypePasswordReset, time.Now()).First(&otpRecord).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired OTP"})
		}

		return c.JSON(fiber.Map{"message": "OTP verified successfully"})
	}
}

// ResetPasswordHandler godoc
// @Summary Reset Password
// @Description Resets the user's password using a valid OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string,otp=string,new_password=string} true "Reset payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/reset-password [post]
func ResetPasswordHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			Email       string `json:"email"`
			OTP         string `json:"otp"`
			NewPassword string `json:"new_password"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// Verify OTP again (and mark used)
		var otpRecord otp.OTP
		if err := db.Where("email = ? AND code = ? AND type = ? AND used = false AND expires_at > ?",
			payload.Email, payload.OTP, otp.TypePasswordReset, time.Now()).First(&otpRecord).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired OTP"})
		}

		// Find User
		var u user.User
		if err := db.Where("email = ?", payload.Email).First(&u).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		// Hash new password
		hashed, err := user.HashPassword(payload.NewPassword)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to process password"})
		}

		// Update User
		u.Password = hashed
		db.Save(&u)

		// Mark OTP used
		otpRecord.Used = true
		db.Save(&otpRecord)

		return c.JSON(fiber.Map{"message": "Password reset successfully"})
	}
}
