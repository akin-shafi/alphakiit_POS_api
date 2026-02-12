package middleware

import (
	"fmt"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/types"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ModuleGuard(db *gorm.DB, module subscription.ModuleType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get user claims
		user, _ := c.Locals("user").(*types.UserClaims)

		if user != nil {
			fmt.Printf("[DEBUG] ModuleGuard - User: %s, Role: %s, Module: %v, Path: %s %s\n", user.UserName, user.Role, module, c.Method(), c.Path())
		}

		// 2. Bypass for super_admin and installers (case-insensitive)
		if user != nil {
			role := strings.TrimSpace(user.Role)
			isSuper := strings.EqualFold(role, "super_admin")
			isInstaller := strings.EqualFold(role, "installer")
			fmt.Printf("[DEBUG] ModuleGuard Bypass Check - Role: '%s', len: %d, isInstaller: %v\n", role, len(role), isInstaller)
			if isSuper || isInstaller {
				fmt.Printf("[DEBUG] ModuleGuard Bypassing for %s\n", role)
				return c.Next()
			}
		}

		// 3. Fallback check for role local
		userRole := c.Locals("role")
		if userRole != nil {
			roleStr, ok := userRole.(string)
			if ok && (strings.EqualFold(roleStr, "super_admin") || strings.EqualFold(roleStr, "installer")) {
				return c.Next()
			}
		}

		val := c.Locals("current_business_id")
		if val == nil {
			role := "unknown"
			if user != nil {
				role = user.Role
			}
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Business context required (Role: %s)", role)})
		}

		businessID := val.(uint)
		if businessID == 0 {
			role := "unknown"
			if user != nil {
				role = user.Role
			}
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Business context required (Role: %s)", role)})
		}

		if !subscription.HasModule(db, businessID, module) {
			return c.Status(403).JSON(fiber.Map{
				"error":   "Module Not Subscribed",
				"module":  module,
				"message": "This feature requires an active subscription for the " + string(module) + " module.",
			})
		}

		return c.Next()
	}
}
