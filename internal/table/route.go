// internal/table/route.go
package table

import (
	"github.com/gofiber/fiber/v2"
)

func (c *TableController) RegisterRoutes(router fiber.Router) {
	router.Get("/tables/sections", c.GetSections)
	router.Post("/tables", c.CreateTable)
	router.Get("/tables", c.ListTables)
	router.Get("/tables/:id", c.GetTable)
	router.Put("/tables/:id", c.UpdateTable)
	router.Delete("/tables/:id", c.DeleteTable)
	router.Get("/tables/:id/orders", c.GetTableOrders)
}
