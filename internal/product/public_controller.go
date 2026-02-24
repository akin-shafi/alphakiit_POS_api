package product

import (
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PublicMenuResponse represents the public catalog data
type PublicMenuResponse struct {
	Business   business.Business   `json:"business"`
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
		var biz business.Business
		if err := db.Where("slug = ?", slug).First(&biz).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fiber.NewError(fiber.StatusNotFound, "business not found")
			}
			return fiber.ErrInternalServerError
		}

		// 2. Fetch Categories
		cats, err := category.ListByBusiness(db, biz.ID)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 3. Fetch Active Products
		products, err := ListByBusiness(db, biz.ID, func(q *gorm.DB) *gorm.DB {
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
