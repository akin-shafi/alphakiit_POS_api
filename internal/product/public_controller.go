package product

import (
	"errors"
	"pos-fiber-app/internal/category"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PublicMenuResponse represents the public catalog data
type PublicMenuResponse struct {
	Business   interface{}         `json:"business"`
	Categories []category.Category `json:"categories"`
	Products   []Product           `json:"products"`
}

// GetPublicMenuBySlugHandler fetches the catalog for a business using its unique slug
func GetPublicMenuBySlugHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slug := c.Params("slug")
		if slug == "" {
			return fiber.NewError(fiber.StatusBadRequest, "slug is required")
		}

		// 1. Find Business
		var biz map[string]interface{}
		if err := db.Table("businesses").Where("slug = ?", slug).Take(&biz).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "business not found")
			}
			return fiber.ErrInternalServerError
		}

		bizID := uint(biz["id"].(int64))

		// 2. Fetch Categories
		cats, err := category.ListByBusiness(db, bizID)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 3. Fetch Active Products
		products, err := ListByBusiness(db, bizID, func(q *gorm.DB) *gorm.DB {
			return q.Where("products.active = ?", true)
		})
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(PublicMenuResponse{
			Business:   biz,
			Categories: cats,
			Products:   products,
		})
	}
}
