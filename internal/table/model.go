// internal/table/model.go
package table

import (
	"time"

	"gorm.io/gorm"
)

type TableStatus string

const (
	StatusAvailable TableStatus = "available"
	StatusOccupied  TableStatus = "occupied"
	StatusReserved  TableStatus = "reserved"
)

// Table represents a physical table or section in a restaurant/bar
type Table struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	BusinessID  uint        `gorm:"index:idx_business_table" json:"business_id"`
	TableNumber string      `gorm:"index:idx_business_table" json:"table_number"` // "T1", "VIP-3", etc
	Section     string      `json:"section,omitempty"`                            // "Outdoor", "VIP", "Main Hall"
	Capacity    int         `json:"capacity" gorm:"default:4"`
	Status      TableStatus `gorm:"type:varchar(20);default:'available'" json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// TableWithOrders extends Table with current orders information
type TableWithOrders struct {
	Table
	ActiveOrders int     `json:"active_orders"`
	TotalAmount  float64 `json:"total_amount"`
}

// Migrate runs the database migration for tables
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Table{})
}
