// internal/product/route.go
package product

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterPublicProductRoutes(r fiber.Router, db *gorm.DB) {
	r.Get("/public/menu/:slug", GetPublicMenuBySlugHandler(db))
}

func RegisterProductRoutes(r fiber.Router, db *gorm.DB) {
	r.Get("/products", ListHandler(db))
	r.Post("/products", CreateHandler(db))
	r.Get("/products/:id", GetHandler(db))
	r.Put("/products/:id", UpdateHandler(db))
	r.Delete("/products/:id", DeleteHandler(db))
}
