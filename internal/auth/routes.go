// internal/auth/route.go
package auth

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterAuthRoutes registers all authentication-related routes
// Public: login, refresh, forgot-password flow
// Protected: logout
func RegisterAuthRoutes(public fiber.Router, protected fiber.Router, db *gorm.DB) {
	// Public routes
	public.Post("/auth/login", Login(db))
	public.Post("/auth/refresh", Refresh(db))

	// Password Reset Flow
	public.Post("/auth/forgot-password", ForgotPasswordHandler(db))
	public.Post("/auth/verify-otp", VerifyOTPHandler(db))
	public.Post("/auth/reset-password", ResetPasswordHandler(db))

	// Protected routes
	protected.Post("/auth/logout", Logout(db))
}
