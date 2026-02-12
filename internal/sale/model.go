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

type PrepStatus string

const (
	PrepPending   PrepStatus = "PENDING"
	PrepPreparing PrepStatus = "PREPARING"
	PrepReady     PrepStatus = "READY"
	PrepServed    PrepStatus = "SERVED"
)

type Sale struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	BusinessID       uint       `gorm:"index;index:idx_business_saledate" json:"business_id"`
	TenantID         string     `gorm:"index;size:8" json:"tenant_id"`
	CustomerName     string     `json:"customer_name,omitempty"`
	CustomerPhone    string     `json:"customer_phone,omitempty"`
	Subtotal         float64    `gorm:"type:decimal(12,2)" json:"subtotal"`
	Tax              float64    `gorm:"type:decimal(12,2)" json:"tax"`
	Discount         float64    `gorm:"type:decimal(12,2)" json:"discount"`
	Total            float64    `gorm:"type:decimal(12,2)" json:"total"`
	PaymentMethod    string     `json:"payment_method"` // CASH, CARD, TRANSFER, etc.
	Status           SaleStatus `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	TerminalID       uint       `json:"terminal_id"`
	OutletID         uint       `gorm:"index" json:"outlet_id"`
	CashierID        uint       `json:"cashier_id"`
	TerminalProvider string     `json:"terminal_provider,omitempty"`              // moniepoint, opay, etc.
	DailySequence    int        `gorm:"type:int;default:0" json:"daily_sequence"` // resets daily
	SaleDate         time.Time  `gorm:"index:idx_business_saledate" json:"sale_date"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"` // for offline sync
	// New fields for table management and shift tracking
	TableID           *uint          `gorm:"index" json:"table_id,omitempty"`
	TableNumber       string         `json:"table_number,omitempty"`                               // Snapshot for history
	OrderType         string         `gorm:"type:varchar(20);default:'dine-in'" json:"order_type"` // dine-in, takeaway, delivery
	ShiftID           *uint          `gorm:"index" json:"shift_id,omitempty"`                      // Link to cashier's shift
	PreparationStatus PrepStatus     `gorm:"type:varchar(20);default:'PENDING'" json:"preparation_status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	CashierName string     `gorm:"-" json:"cashier_name"` // Populated manually or via join
	SaleItems   []SaleItem `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE" json:"items"`
}

type SaleItem struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	SaleID            uint       `gorm:"index" json:"sale_id"`
	ProductID         uint       `json:"product_id"`
	ProductName       string     `json:"product_name"` // snapshot
	Quantity          int        `json:"quantity"`
	UnitPrice         float64    `gorm:"type:decimal(12,2)" json:"unit_price"` // snapshot
	TotalPrice        float64    `gorm:"type:decimal(12,2)" json:"total_price"`
	PreparationStatus PrepStatus `gorm:"type:varchar(20);default:'PENDING'" json:"preparation_status"`
}

type SalesReport struct {
	FromDate                     string  `json:"from_date"`
	ToDate                       string  `json:"to_date"`
	TotalSales                   float64 `json:"total_sales"`
	TotalTransactions            int     `json:"total_transactions"`
	CashSales                    float64 `json:"cash_sales"`
	CashTransactions             int     `json:"cash_transactions"`
	CardSales                    float64 `json:"card_sales"`
	CardTransactions             int     `json:"card_transactions"`
	TransferSales                float64 `json:"transfer_sales"`
	TransferTransactions         int     `json:"transfer_transactions"`
	MobileMoneySales             float64 `json:"mobile_money_sales"`
	MobileMoneyTransactions      int     `json:"mobile_money_transactions"`
	ExternalTerminalSales        float64 `json:"external_terminal_sales"`
	ExternalTerminalTransactions int     `json:"external_terminal_transactions"`
	CreditSales                  float64 `json:"credit_sales"`
	CreditTransactions           int     `json:"credit_transactions"`
	OtherSales                   float64 `json:"other_sales"`
	OtherTransactions            int     `json:"other_transactions"`
	AverageSale                  float64 `json:"average_sale"`
}

// SaleSummary stores daily aggregates for data older than the retention period
type SaleSummary struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	BusinessID            uint      `gorm:"index;uniqueIndex:idx_biz_date" json:"business_id"`
	Date                  time.Time `gorm:"uniqueIndex:idx_biz_date" json:"date"`
	TotalSales            float64   `gorm:"type:decimal(12,2)" json:"total_sales"`
	TotalTransactions     int       `json:"total_transactions"`
	CashSales             float64   `gorm:"type:decimal(12,2)" json:"cash_sales"`
	CardSales             float64   `gorm:"type:decimal(12,2)" json:"card_sales"`
	TransferSales         float64   `gorm:"type:decimal(12,2)" json:"transfer_sales"`
	ExternalTerminalSales float64   `gorm:"type:decimal(12,2)" json:"external_terminal_sales"`
	CreditSales           float64   `gorm:"type:decimal(12,2)" json:"credit_sales"`
	Tax                   float64   `gorm:"type:decimal(12,2)" json:"tax"`
	Discount              float64   `gorm:"type:decimal(12,2)" json:"discount"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
