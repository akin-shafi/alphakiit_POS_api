package recipe

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRecipeRoutes(router fiber.Router, db *gorm.DB) {
	group := router.Group("/recipes")

	group.Get("/:product_id", GetRecipeHandler(db))
	group.Post("/", AddIngredientHandler(db))
	group.Delete("/:id", RemoveIngredientHandler(db))
}
