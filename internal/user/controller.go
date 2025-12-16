package user

import (
	"pos-fiber-app/internal/types"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// --------------------- Create User ---------------------

// @Summary Create user
// @Description Create a user under a tenant
// @Tags Users
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body User true "User payload"
// @Success 201 {object} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func CreateUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing X-Tenant-ID"})
		}

		user := new(User)
		if err := c.BodyParser(user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		user.TenantID = tenantID

		if err := service.Create(user); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(201).JSON(user)
	}
}

// --------------------- List Users ---------------------

// @Summary List users
// @Description List users for a tenant
// @Tags Users
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {array} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [get]
func ListUsersHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing X-Tenant-ID"})
		}

		users, err := service.ListByTenant(tenantID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(users)
	}
}

// --------------------- Get User Profile ---------------------

// @Summary Get user profile
// @Description Retrieve profile of a user
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func GetUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		user, err := service.GetByID(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		return c.JSON(user)
	}
}

// --------------------- Update User ---------------------

// @Summary Update user
// @Description Update user details
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body User true "User payload"
// @Success 200 {object} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func UpdateUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		user := new(User)
		if err := c.BodyParser(user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		updatedUser, err := service.Update(id, user)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(updatedUser)
	}
}

// --------------------- Delete User ---------------------

// @Summary Delete user
// @Description Delete a user by ID
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func DeleteUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		if err := service.Delete(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendStatus(204)
	}
}

// --------------------- Reset Password ---------------------

// @Summary Reset user password
// @Description Reset a user's password
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body map[string]string true "Password payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/reset-password [post]
func ResetPasswordHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}

		var payload struct {
			Password string `json:"password"`
		}

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		if err := service.ResetPassword(uint(id), payload.Password); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "Password reset successfully"})
	}
}

// ProfileHandler retrieves the logged-in user's profile
// @Summary Get profile
// @Description Get profile of logged-in user
// @Tags Users
// @Success 200 {object} User
// @Security BearerAuth
// @Router /profile [get]
func ProfileHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*types.UserClaims) // JWT claims from middleware
		u, err := service.GetByID(userCtx.UserID)       // UserID is uint
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.JSON(u)
	}
}
