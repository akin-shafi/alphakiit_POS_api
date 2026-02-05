package shift

import (
	"github.com/gofiber/fiber/v2"
)

func (c *ShiftController) RegisterRoutes(router fiber.Router) {
	router.Post("/shifts/start", c.StartShift)
	router.Post("/shifts/:id/end", c.EndShift)
	router.Get("/shifts/active", c.GetActiveShift)
	router.Get("/shifts/:id/summary", c.GetShiftSummary)
	router.Get("/shifts", c.ListShifts)
}
