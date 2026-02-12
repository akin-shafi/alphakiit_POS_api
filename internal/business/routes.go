package business

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	// "gorm.io/gorm"
)

func RegisterBusinessRoutes(r fiber.Router, db *gorm.DB) {
	r.Get("/businesses", ListHandler(db))
	r.Post("/businesses", CreateHandler(db))
	r.Get("/businesses/:id", GetHandler(db))
	r.Put("/businesses/:id", UpdateHandler(db))
	r.Delete("/businesses/:id", DeleteHandler(db))

	// Google Drive Connect
	r.Get("/google/connect", ConnectGoogleDriveHandler(db))
}

func RegisterPublicBusinessRoutes(r fiber.Router, db *gorm.DB) {
	r.Get("/google/callback", GoogleCallbackHandler(db))
}
