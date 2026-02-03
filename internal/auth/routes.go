// internal/auth/route.go
package auth

import (
	"pos-fiber-app/internal/config"
	"pos-fiber-app/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(router fiber.Router, db *gorm.DB) {
	// Public routes with LoginLimiter
	router.Post("/login", config.LoginLimiter(), Login(db))

	// Other Public routes
	router.Post("/refresh", Refresh(db))
	router.Post("/verify-email", VerifyEmail(db))
	router.Post("/resend-otp", ResendOTP(db))

	// Password Reset Flow (Public)
	router.Post("/forgot-password", ForgotPasswordHandler(db))
	router.Post("/verify-reset-otp", VerifyResetOTPHandler(db))
	router.Post("/reset-password", ResetPasswordHandler(db))

	// Protected routes
	router.Post("/logout", middleware.JWTProtected(), Logout(db))
}
