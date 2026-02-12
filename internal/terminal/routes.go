package terminal

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	r.Post("/terminals", Register(db))

	// Printer Management
	r.Post("/printers", AddPrinter(db))
	r.Get("/printers", ListPrinters(db))
	r.Put("/printers/:id", UpdatePrinter(db))
	r.Delete("/printers/:id", DeletePrinter(db))
}
