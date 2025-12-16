// internal/category/route.go
package category

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterCategoryRoutes(r fiber.Router, db *gorm.DB) {
	r.Get("/categories", ListHandler(db))
	r.Post("/categories", CreateHandler(db))
	r.Get("/categories/:id", GetHandler(db))
	r.Put("/categories/:id", UpdateHandler(db))
	r.Delete("/categories/:id", DeleteHandler(db))
}
