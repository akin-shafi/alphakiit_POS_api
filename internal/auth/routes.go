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
	router.Post("/installer/register", RegisterInstaller(db))

	// Other Public routes
	router.Post("/refresh", Refresh(db))
	router.Post("/verify-email", VerifyEmail(db))
	router.Post("/resend-otp", ResendOTP(db))

	// Password Reset Flow (Public)
	router.Post("/forgot-password", ForgotPasswordHandler(db))
	router.Post("/verify-reset-otp", VerifyResetOTPHandler(db))
	router.Post("/reset-password", ResetPasswordHandler(db))

	// Protected routes
	protected := router.Group("/", middleware.JWTProtected())
	protected.Post("/logout", Logout(db))
	protected.Get("/profile", ProfileHandler(db))
	protected.Get("/sessions", GetActiveSessions(db))
	protected.Delete("/sessions/logout-all", LogoutAllUserSessions(db))
	protected.Delete("/sessions/:id", RevokeSession(db))
}
