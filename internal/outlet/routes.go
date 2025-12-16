package outlet

import (
	"pos-fiber-app/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	r.Post("/outlets", middleware.RequireRoles("OWNER", "MANAGER"), Create(db))
}
