package user

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router fiber.Router, db *gorm.DB) {
	service := NewUserService(db)

	// Create user
	router.Post("/users", CreateUserHandler(service))

	// List users for tenant
	router.Get("/users", ListUsersHandler(service))

	// Get single user
	router.Get("/users/:id", GetUserHandler(service))

	// Update user
	router.Put("/users/:id", UpdateUserHandler(service))

	// Delete user
	router.Delete("/users/:id", DeleteUserHandler(service))

	// Reset user password
	router.Post("/users/:id/reset-password", ResetPasswordHandler(service))

	// Get logged-in user profile
	router.Get("/profile", ProfileHandler(service))
}
