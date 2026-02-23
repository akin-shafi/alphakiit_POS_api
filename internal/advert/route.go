package advert

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterAdvertRoutes(router fiber.Router, db *gorm.DB) {
	controller := NewController(db)

	group := router.Group("/adverts")
	group.Post("/", controller.CreateAdvert)
	group.Get("/", controller.GetAdverts)
	group.Put("/:id", controller.UpdateAdvert)
	group.Delete("/:id", controller.DeleteAdvert)
}

func RegisterPublicAdvertRoutes(router fiber.Router, db *gorm.DB) {
	controller := NewController(db)

	group := router.Group("/public/adverts")
	group.Get("/", controller.GetAdverts)
}
