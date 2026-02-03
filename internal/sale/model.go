// internal/sale/model.go
package sale

import (
	"time"

	"gorm.io/gorm"
	// "pos-fiber-app/internal/business"
	// "pos-fiber-app/internal/common"
)

type SaleStatus string

const (
	StatusDraft     SaleStatus = "DRAFT"
	StatusCompleted SaleStatus = "COMPLETED"
	StatusVoided    SaleStatus = "VOIDED"
	StatusHeld      SaleStatus = "HELD" // parked for later
)

type Sale struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	BusinessID    uint           `gorm:"index" json:"business_id"`
	TenantID      string         `gorm:"index;size:8" json:"tenant_id"`
	CustomerName  string         `json:"customer_name,omitempty"`
	CustomerPhone string         `json:"customer_phone,omitempty"`
	Subtotal      float64        `gorm:"type:decimal(12,2)" json:"subtotal"`
	Tax           float64        `gorm:"type:decimal(12,2)" json:"tax"`
	Discount      float64        `gorm:"type:decimal(12,2)" json:"discount"`
	Total         float64        `gorm:"type:decimal(12,2)" json:"total"`
	PaymentMethod string         `json:"payment_method"` // CASH, CARD, TRANSFER, etc.
	Status        SaleStatus     `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	TerminalID    uint           `json:"terminal_id"`
	CashierID     uint           `json:"cashier_id"`
	DailySequence int            `gorm:"type:int;default:0" json:"daily_sequence"` // resets daily
	SaleDate      time.Time      `json:"sale_date"`
	SyncedAt      *time.Time     `json:"synced_at,omitempty"` // for offline sync
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	CashierName string     `gorm:"-" json:"cashier_name"` // Populated manually or via join
	SaleItems   []SaleItem `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE" json:"items"`
}

type SaleItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	SaleID      uint    `gorm:"index" json:"sale_id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"` // snapshot
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `gorm:"type:decimal(12,2)" json:"unit_price"` // snapshot
	TotalPrice  float64 `gorm:"type:decimal(12,2)" json:"total_price"`
}

type SalesReport struct {
	FromDate                string  `json:"from_date"`
	ToDate                  string  `json:"to_date"`
	TotalSales              float64 `json:"total_sales"`
	TotalTransactions       int     `json:"total_transactions"`
	CashSales               float64 `json:"cash_sales"`
	CashTransactions        int     `json:"cash_transactions"`
	CardSales               float64 `json:"card_sales"`
	CardTransactions        int     `json:"card_transactions"`
	TransferSales           float64 `json:"transfer_sales"`
	TransferTransactions    int     `json:"transfer_transactions"`
	MobileMoneySales        float64 `json:"mobile_money_sales"`
	MobileMoneyTransactions int     `json:"mobile_money_transactions"`
	OtherSales              float64 `json:"other_sales"`
	OtherTransactions       int     `json:"other_transactions"`
	AverageSale             float64 `json:"average_sale"`
}
