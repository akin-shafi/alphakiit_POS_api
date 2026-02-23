package advert

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Controller struct {
	service *Service
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{
		service: NewService(db),
	}
}

// CreateAdvert handles POST /adverts
func (ctrl *Controller) CreateAdvert(c *fiber.Ctx) error {
	var req CreateAdvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	advert, err := ctrl.service.CreateAdvert(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(advert)
}

// GetAdverts handles GET /adverts
func (ctrl *Controller) GetAdverts(c *fiber.Ctx) error {
	businessIDStr := c.Query("business_id")
	var businessIDPtr *uint
	if businessIDStr != "" {
		id, err := strconv.ParseUint(businessIDStr, 10, 32)
		if err == nil {
			uID := uint(id)
			businessIDPtr = &uID
		}
	}

	adverts, err := ctrl.service.GetAdverts(businessIDPtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(adverts)
}

// UpdateAdvert handles PUT /adverts/:id
func (ctrl *Controller) UpdateAdvert(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid advert id"})
	}

	var req UpdateAdvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	advert, err := ctrl.service.UpdateAdvert(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(advert)
}

// DeleteAdvert handles DELETE /adverts/:id
func (ctrl *Controller) DeleteAdvert(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid advert id"})
	}

	if err := ctrl.service.DeleteAdvert(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
