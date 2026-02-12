package printing

import (
	"fmt"
	"pos-fiber-app/internal/subscription"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

// PrintWebsocketHandler handles the connection for local print agents
func PrintWebsocketHandler(db *gorm.DB) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		outletIDVal := conn.Query("outlet_id")
		if outletIDVal == "" {
			conn.WriteJSON(fiber.Map{"error": "missing outlet id"})
			conn.Close()
			return
		}

		outletID, _ := strconv.ParseUint(outletIDVal, 10, 64)
		bizIDVal := conn.Locals("current_business_id")
		if bizIDVal == nil {
			conn.WriteJSON(fiber.Map{"error": "missing business id"})
			conn.Close()
			return
		}

		bizID := bizIDVal.(uint)

		// Check if module is subscribed (Optional: create a 'PRINTING' module if needed)
		// For now we use KDS as a proxy for premium features or just allow it
		if !subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			// conn.WriteJSON(fiber.Map{"error": "Printing feature requires KDS or PRO module"})
			// conn.Close()
			// return
		}

		agentConn := &AgentConn{
			OutletID: uint(outletID),
			Conn:     conn,
		}

		GlobalPrintingHub.Register <- agentConn

		defer func() {
			GlobalPrintingHub.Unregister <- agentConn
		}()

		// Keep connection alive/listen for heartbeats
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Respond to pings
			if string(msg) == "ping" {
				conn.WriteMessage(mt, []byte("pong"))
			}
		}
	})
}

// TestPrintHandler allows testing the connection
func TestPrintHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		outletID, _ := strconv.ParseUint(c.Query("outlet_id"), 10, 64)
		printerID, _ := strconv.ParseUint(c.Query("printer_id"), 10, 64)

		if outletID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "outlet_id is required"})
		}

		job := PrintJob{
			PrinterID: uint(printerID),
			Content:   fmt.Sprintf("TEST PRINT\nOutlet: %d\nTime: 2026-02-07\n\n\n\x1dV\x00", outletID),
		}

		GlobalPrintingHub.SendJobToOutlet(uint(outletID), job)

		return c.JSON(fiber.Map{"status": "queued", "message": "Test job sent to outlet agents"})
	}
}
