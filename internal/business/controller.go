package business

import (
	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListBusinesses godoc
// @Summary List all businesses for tenant
// @Tags Business
// @Security BearerAuth
// @Success 200 {array} Business
// @Router /businesses [get]
func ListHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)
		businesses, err := ListBusinesses(db, claims.TenantID)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(businesses)
	}
}

// CreateBusiness godoc
// @Summary Create a new business/outlet
// @Description Create a new business with selected type and currency
// @Tags Business
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateBusinessRequest true "Business information including currency"
// @Success 201 {object} Business
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /businesses [post]
func CreateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateBusinessRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		claims := c.Locals("user").(*types.UserClaims)

		biz := &Business{
			TenantID: claims.TenantID,
			Name:     req.Name,
			Type:     req.Type,
			Address:  req.Address,
			City:     req.City,
			Currency: req.Currency,
		}

		if err := db.Create(biz).Error; err != nil {
			return fiber.ErrInternalServerError
		}

		// Optional: Trigger seeding here
		SeedNewBusiness(db, biz) // Synchronous, safe, fast

		// Or fire-and-forget if you prefer:
		// go SeedNewBusiness(db, biz)
		return c.Status(fiber.StatusCreated).JSON(biz)
	}
}

// GetBusiness godoc
// @Summary Get specific business
// @Tags Business
// @Security BearerAuth
// @Param id path uint true "Business ID"
// @Success 200 {object} Business
// @Router /businesses/{id} [get]
func GetHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, _ := c.ParamsInt("id")
		claims := c.Locals("user").(*types.UserClaims)

		biz, err := GetBusiness(db, uint(id), claims.TenantID)
		if err != nil {
			return fiber.NewError(404, err.Error())
		}
		return c.JSON(biz)
	}
}

// UpdateHandler godoc
// @Summary Update a specific business
// @Description Update name, type, address or other details of a business belonging to the authenticated user's tenant
// @Tags Business
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Business ID"
// @Param body body UpdateBusinessRequest true "Business update payload"
// @Success 200 {object} Business
// @Failure 400 {object} map[string]string "Invalid payload or validation error"
// @Failure 404 {object} map[string]string "Business not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /businesses/{id} [put]
func UpdateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get business ID from path
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid business ID")
		}

		// Parse request body
		var req UpdateBusinessRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		// Basic validation
		if req.Name == "" && req.Type == "" && req.Address == "" && req.City == "" {
			return fiber.NewError(fiber.StatusBadRequest, "at least one field must be provided for update")
		}

		// Get tenant from JWT claims
		claims := c.Locals("user").(*types.UserClaims)

		// Prepare updates map (only non-empty fields)
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Type != "" {
			updates["type"] = req.Type
		}
		if req.Address != "" {
			updates["address"] = req.Address
		}
		if req.City != "" {
			updates["city"] = req.City
		}

		if req.Currency != nil {
			updates["currency"] = *req.Currency
		}

		// Perform update
		biz, err := UpdateBusiness(db, uint(id), claims.TenantID, updates)
		if err != nil {
			if err.Error() == "business not found" {
				return fiber.NewError(fiber.StatusNotFound, "business not found")
			}
			return fiber.ErrInternalServerError
		}

		return c.JSON(biz)
	}
}

// DeleteHandler godoc
// @Summary Delete a specific business
// @Description Soft delete a business (and potentially all related data like outlets, terminals, products)
// @Tags Business
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Business ID"
// @Success 204 "Business deleted successfully"
// @Failure 404 {object} map[string]string "Business not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /businesses/{id} [delete]
func DeleteHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get business ID from path
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid business ID")
		}

		// Get tenant from claims
		claims := c.Locals("user").(*types.UserClaims)

		// Perform delete
		if err := DeleteBusiness(db, uint(id), claims.TenantID); err != nil {
			if err.Error() == "business not found" {
				return fiber.NewError(fiber.StatusNotFound, "business not found")
			}
			return fiber.ErrInternalServerError
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
