package terminal

import (
	"time"
)

type PrinterType string

const (
	PrinterReceipt PrinterType = "RECEIPT"
	PrinterKitchen PrinterType = "KITCHEN"
	PrinterBar     PrinterType = "BAR"
)

type Printer struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	TenantID  string      `gorm:"index" json:"tenant_id"`
	OutletID  uint        `gorm:"index" json:"outlet_id"`
	Name      string      `json:"name"`                              // e.g. "Kitchen Thermal"
	Type      PrinterType `gorm:"type:varchar(20)" json:"type"`      // e.g. "KITCHEN"
	Interface string      `gorm:"type:varchar(20)" json:"interface"` // e.g. "USB", "NETWORK"
	Address   string      `json:"address"`                           // e.g. "192.168.1.100" or USB device name
	IsActive  bool        `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
