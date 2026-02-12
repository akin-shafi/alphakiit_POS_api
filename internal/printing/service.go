package printing

import (
	"fmt"
	"pos-fiber-app/internal/terminal"

	"gorm.io/gorm"
)

// PrintKitchenOrder formats and sends an order to kitchen printers in an outlet
func PrintKitchenOrder(db *gorm.DB, tenantID string, outletID uint, orderData interface{}) error {
	// 1. Find all kitchen printers for this outlet
	var printers []terminal.Printer
	if err := db.Where("tenant_id = ? AND outlet_id = ? AND type = ?", tenantID, outletID, terminal.PrinterKitchen).Find(&printers).Error; err != nil {
		return err
	}

	if len(printers) == 0 {
		return nil // No kitchen printers configured
	}

	// 2. Format the order (Simplified ESC/POS generation)
	content := formatOrderForKitchen(orderData)

	// 3. Send to each printer via connected agents
	for _, p := range printers {
		job := PrintJob{
			PrinterID: p.ID,
			Content:   content,
			Data:      orderData,
		}
		GlobalPrintingHub.SendJobToOutlet(outletID, job)
	}

	return nil
}

func formatOrderForKitchen(data interface{}) string {
	// This would be much more complex in a real app (looping through items, bolding names, etc.)
	// For now, it's a simple text block with a cut command at the end.

	header := "\x1b\x61\x01" + // Center alignment
		"\x1b\x21\x30" + // Double height/width
		"KITCHEN ORDER\n" +
		"\x1b\x21\x00" + // Reset font
		"--------------------------------\n"

	body := fmt.Sprintf("Order ID: %v\n", data) // In reality, cast and loop

	footer := "\n\n\n\x1dV\x00" // Form feed and cut

	return header + body + footer
}

// Helper to check if printing is enabled for a business/terminal
// (This could be expanded to check specific database settings)
func IsSilentPrintingEnabled(db *gorm.DB, bizID uint) bool {
	// For now, always return true if you have the module,
	// but you could check a specific 'silent_printing' flag in the Business model.
	return true
}
