package tutorial

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(api fiber.Router, db *gorm.DB) {
	service := NewTutorialService(db)
	controller := NewTutorialController(service)

	tutorials := api.Group("/tutorials")
	tutorials.Get("/", controller.GetTutorials)
}
