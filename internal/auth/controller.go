package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/types"
	"pos-fiber-app/internal/user"
)

// Role type
type Role string

// Claims represents JWT token claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
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
		if err := db.Where("LOWER(email) = ?", strings.ToLower(req.Email)).First(&u).Error; err != nil {
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
			Email:     strings.ToLower(req.Email),
			Code:      code,
			Type:      otp.TypeVerification,
			ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
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
			"LOWER(email) = ? AND code = ? AND type = ? AND used = false AND expires_at > ?",
			strings.ToLower(req.Email),
			req.Code,
			otp.TypeVerification,
			time.Now().UTC(),
		).First(&otpRecord).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid or expired OTP",
			})
		}

		var u user.User
		if err := db.Where("LOWER(email) = ?", strings.ToLower(req.Email)).First(&u).Error; err != nil {
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
			UserName: u.FirstName + " " + u.LastName,
			TenantID: u.TenantID,
			Role:     u.Role,
			OutletID: u.OutletID,
		}

		access, _ := GenerateAccessToken(claims)
		refresh, _ := GenerateRefreshToken(claims)

		db.Create(&RefreshToken{
			UserID:     u.ID,
			Token:      refresh,
			DeviceInfo: c.Get("User-Agent"),
			IPAddress:  c.IP(),
			LastUsedAt: time.Now(),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		})

		var biz business.Business
		var tnt business.Tenant

		if u.TenantID != "" {
			db.Where("tenant_id = ?", u.TenantID).First(&biz)
			db.Where("id = ?", u.TenantID).First(&tnt)
		}

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

// RegisterInstallerRequest payload
type RegisterInstallerRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// RegisterInstaller godoc
func RegisterInstaller(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterInstallerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
		}

		// Check if user exists
		var existing user.User
		if err := db.Where("email = ?", strings.ToLower(req.Email)).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email already registered"})
		}

		hashed, _ := user.HashPassword(req.Password)

		u := user.User{
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			Email:      strings.ToLower(req.Email),
			Password:   hashed,
			Role:       "INSTALLER",
			IsVerified: false,
			Active:     true,
		}

		if err := db.Create(&u).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create account"})
		}

		// Send Verification OTP
		code, _ := otp.GenerateOTP()
		otpEntry := otp.OTP{
			Email:     strings.ToLower(u.Email),
			Code:      code,
			Type:      otp.TypeVerification,
			ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			Used:      false,
		}
		db.Create(&otpEntry)

		sender := email.NewSender(email.LoadConfig())
		body := fmt.Sprintf("<h2>Verify Your Email</h2><p>Your OTP code is <b>%s</b></p>", code)
		go sender.SendCustomEmail(u.Email, "Verify Your Email", body)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "account created, please verify email",
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

		if !user.CheckPassword(u.Password, req.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid credentials",
			})
		}

		if !u.IsVerified {
			// Generate OTP for verification
			code, _ := otp.GenerateOTP()
			otpEntry := otp.OTP{
				Email:     strings.ToLower(u.Email),
				Code:      code,
				Type:      otp.TypeVerification,
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour), // Give them a day for first set up
				Used:      false,
			}
			db.Create(&otpEntry)

			sender := email.NewSender(email.LoadConfig())
			body := fmt.Sprintf(
				"<h2>Email Verification</h2><p>Your OTP code is <b>%s</b></p><p>This code expires in 15 minutes.</p>",
				code,
			)
			go sender.SendCustomEmail(u.Email, "Verify Your Email", body)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "please verify your email address",
			})
		}

		claims := Claims{
			UserID:   u.ID,
			UserName: u.FirstName + " " + u.LastName,
			TenantID: u.TenantID,
			Role:     u.Role,
			OutletID: u.OutletID,
		}

		access, _ := GenerateAccessToken(claims)
		refresh, _ := GenerateRefreshToken(claims)

		db.Create(&RefreshToken{
			UserID:     u.ID,
			Token:      refresh,
			DeviceInfo: c.Get("User-Agent"),
			IPAddress:  c.IP(),
			LastUsedAt: time.Now(),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		})

		var biz business.Business
		var tnt business.Tenant

		// Only fetch business/tenant if user belongs to one
		if u.TenantID != "" {
			db.Where("tenant_id = ?", u.TenantID).First(&biz)
			db.Where("id = ?", u.TenantID).First(&tnt)

			// Populate active modules
			if biz.ID != 0 {
				db.Table("business_modules").
					Where("business_id = ? AND is_active = ?", biz.ID, true).
					Pluck("module", &biz.ActiveModules)
			}
		}

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

// Logout invalidates the user's current session
func Logout(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// We could potentially black-list the access token,
		// but for now we just return SUCCESS.
		// If using refresh tokens, we would delete the refresh token here.
		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
			"success": true,
			"message": "logged out",
		})
	}
}

// GetActiveSessions returns all active refresh tokens for the business
func GetActiveSessions(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*types.UserClaims)

		var sessions []struct {
			RefreshToken
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}
		// Join with users to filter by TenantID (business)
		err := db.Table("refresh_tokens").
			Select("refresh_tokens.*, users.first_name, users.last_name, users.email").
			Joins("JOIN users ON users.id = refresh_tokens.user_id").
			Where("users.tenant_id = ? AND refresh_tokens.expires_at > ?", userClaims.TenantID, time.Now()).
			Order("refresh_tokens.last_used_at DESC").
			Scan(&sessions).Error

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    sessions,
		})
	}
}

// LogoutAllUserSessions invalidates all sessions for all users in a tenant
func LogoutAllUserSessions(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*types.UserClaims)

		// Delete all refresh tokens for users in this tenant
		err := db.Exec(`
			DELETE FROM refresh_tokens 
			WHERE user_id IN (SELECT id FROM users WHERE tenant_id = ?)
		`, userClaims.TenantID).Error

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout all users"})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "all users logged out successfully",
		})
	}
}

// RevokeSession invalidates a specific refresh token session
func RevokeSession(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		userClaims := c.Locals("user").(*types.UserClaims)

		// Verify session belongs to this tenant
		var session RefreshToken
		err := db.Table("refresh_tokens").
			Joins("JOIN users ON users.id = refresh_tokens.user_id").
			Where("refresh_tokens.id = ? AND users.tenant_id = ?", id, userClaims.TenantID).
			First(&session).Error

		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
		}

		db.Delete(&session)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "session revoked",
		})
	}
}

// ProfileHandler returns the current logged in user and business info
func ProfileHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*types.UserClaims)
		var u user.User
		if err := db.First(&u, userClaims.UserID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}

		var biz business.Business
		var tnt business.Tenant

		if u.TenantID != "" {
			db.Where("tenant_id = ?", u.TenantID).First(&biz)
			db.Where("id = ?", u.TenantID).First(&tnt)

			// Populate active modules
			if biz.ID != 0 {
				db.Table("business_modules").
					Where("business_id = ? AND is_active = ?", biz.ID, true).
					Pluck("module", &biz.ActiveModules)
			}
		}

		return c.JSON(fiber.Map{
			"user":     u,
			"business": biz,
			"tenant":   tnt,
		})
	}
}
