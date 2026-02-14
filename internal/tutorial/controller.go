package tutorial

import (
	"pos-fiber-app/internal/common"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TutorialController struct {
	service *TutorialService
}

func NewTutorialController(service *TutorialService) *TutorialController {
	return &TutorialController{service: service}
}

func (c *TutorialController) GetTutorials(ctx *fiber.Ctx) error {
	businessTypeStr, _ := ctx.Locals("business_type").(string)

	businessType := common.BusinessType(strings.ToUpper(businessTypeStr))
	if businessType == "" {
		// Fallback for testing or public access if allowed
		businessType = common.BusinessType(strings.ToUpper(ctx.Query("type", "RETAIL")))
	}

	tutorials, err := c.service.GetTutorialsByBusinessType(businessType)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    tutorials,
	})
}
