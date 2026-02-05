package user

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(public fiber.Router, protected fiber.Router, db *gorm.DB) {
	service := NewUserService(db)

	// Public: password reset (some parts might be public)
	// Actually most user management should be protected except maybe self-registration if allowed

	// Create user (Public/Internal depending on flow, but let's keep it in public for now if onboarding uses it)
	public.Post("/users", CreateUserHandler(service))

	// Protected user routes
	userGroup := protected.Group("/users")
	userGroup.Get("/", ListUsersHandler(service))
	userGroup.Get("/:id", GetUserHandler(service))
	userGroup.Put("/:id", UpdateUserHandler(service))
	userGroup.Delete("/:id", DeleteUserHandler(service))
	userGroup.Post("/:id/reset-password", ResetPasswordHandler(service))

	// Logged-in user profile
	protected.Get("/profile", ProfileHandler(service))
}
