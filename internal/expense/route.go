package expense

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	group := r.Group("/expenses")
	group.Post("/", CreateHandler(db))
	group.Get("/", ListHandler(db))
	group.Delete("/:id", DeleteHandler(db))
}
