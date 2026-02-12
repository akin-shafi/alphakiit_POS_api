package seed

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB) {
	router.Post("/seed", SeedHandler(db))
}

func RegisterPublicRoutes(router fiber.Router, db *gorm.DB) {
	router.Post("/seed/installers", SeedInstallerHandler(db))
}
