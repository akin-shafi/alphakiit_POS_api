package printing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	// WebSocket endpoint for Agents
	r.Get("/ws/printing", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, PrintWebsocketHandler(db))

	// API to trigger print jobs (Protected by usual business/outlet middleware)
	r.Post("/print/test", TestPrintHandler(db))
}
