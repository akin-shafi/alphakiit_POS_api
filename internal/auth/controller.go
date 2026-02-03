package auth

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/user"
)

// Role type
type Role string

// Claims represents JWT token claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	OutletID *uint  `json:"outlet_id,omitempty"`
	jwt.RegisteredClaims
}

/*
|--------------------------------------------------------------------------
| RESEND OTP
|--------------------------------------------------------------------------
*/

// ResendOTPRequest payload
type ResendOTPRequest struct {
	Email string `json:"email"`
}

// ResendOTP godoc
// @Summary Resend Verification OTP
// @Description Generates a new OTP and sends it to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body ResendOTPRequest true "Email payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/resend-otp [post]
func ResendOTP(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ResendOTPRequest
		if err := c.BodyParser(&req); err != nil || req.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		var u user.User
		if err := db.Where("email = ?", req.Email).First(&u).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		if u.IsVerified {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "user is already verified",
			})
		}

		code, err := otp.GenerateOTP()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate OTP",
			})
		}

		otpEntry := otp.OTP{
			Email:     req.Email,
			Code:      code,
			Type:      otp.TypeVerification,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			Used:      false,
		}

		if err := db.Create(&otpEntry).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save OTP",
			})
		}

		sender := email.NewSender(email.LoadConfig())
		body := fmt.Sprintf(
			"<h2>Email Verification</h2><p>Your OTP code is <b>%s</b></p><p>This code expires in 15 minutes.</p>",
			code,
		)

		go func() {
			_ = sender.SendCustomEmail(req.Email, "Verify Your Email", body)
		}()

		return c.JSON(fiber.Map{
			"message": "verification code sent",
		})
	}
}

/*
|--------------------------------------------------------------------------
| VERIFY EMAIL
|--------------------------------------------------------------------------
*/

// VerifyEmailRequest payload
type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// VerifyEmail godoc
// @Summary Verify Email OTP
// @Description Verifies the OTP and activates the user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body VerifyEmailRequest true "Verification payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/verify-email [post]
func VerifyEmail(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req VerifyEmailRequest
		if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		var otpRecord otp.OTP
		if err := db.Where(
			"email = ? AND code = ? AND type = ? AND used = false AND expires_at > ?",
			req.Email,
			req.Code,
			otp.TypeVerification,
			time.Now(),
		).First(&otpRecord).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid or expired OTP",
			})
		}

		var u user.User
		if err := db.Where("email = ?", req.Email).First(&u).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		u.IsVerified = true
		u.Active = true
		if err := db.Save(&u).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update user",
			})
		}

		otpRecord.Used = true
		_ = db.Save(&otpRecord)

		claims := Claims{
			UserID:   u.ID,
			TenantID: u.TenantID,
			Role:     u.Role,
			OutletID: u.OutletID,
		}

		access, _ := GenerateAccessToken(claims)
		refresh, _ := GenerateRefreshToken(claims)

		db.Create(&RefreshToken{
			UserID:    u.ID,
			Token:     refresh,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})

		var biz business.Business
		db.Where("tenant_id = ?", u.TenantID).First(&biz)

		var tnt business.Tenant
		db.Where("id = ?", u.TenantID).First(&tnt)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":       "email verified successfully",
			"access_token":  access,
			"refresh_token": refresh,
			"user":          u,
			"business":      biz,
			"tenant":        tnt,
		})
	}
}

/*
|--------------------------------------------------------------------------
| LOGIN
|--------------------------------------------------------------------------
*/

// LoginRequest payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login godoc
// @Summary User Login
// @Description Authenticate user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		var u user.User
		if err := db.Where("email = ? AND active = true", req.Email).First(&u).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid credentials",
			})
		}

		if !u.IsVerified {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "please verify your email address",
			})
		}

		if !user.CheckPassword(u.Password, req.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid credentials",
			})
		}

		claims := Claims{
			UserID:   u.ID,
			TenantID: u.TenantID,
			Role:     u.Role,
			OutletID: u.OutletID,
		}

		access, _ := GenerateAccessToken(claims)
		refresh, _ := GenerateRefreshToken(claims)

		db.Create(&RefreshToken{
			UserID:    u.ID,
			Token:     refresh,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})

		var biz business.Business
		db.Where("tenant_id = ?", u.TenantID).First(&biz)

		var tnt business.Tenant
		db.Where("id = ?", u.TenantID).First(&tnt)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"access_token":  access,
			"refresh_token": refresh,
			"user":          u,
			"business":      biz,
			"tenant":        tnt,
		})
	}
}

/*
|--------------------------------------------------------------------------
| REFRESH & LOGOUT
|--------------------------------------------------------------------------
*/

// Refresh godoc
func Refresh(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}

// Logout godoc
func Logout(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	}
}
