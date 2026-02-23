package business

import (
	"time"
)

// TrialChecklist tracks the activation progress for a business trial
type TrialChecklist struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	BusinessID uint `gorm:"uniqueIndex" json:"business_id"`

	// Day 0 - Account Setup
	BusinessInfoCompleted bool `gorm:"default:false" json:"business_info_completed"`
	DeviceConnected       bool `gorm:"default:false" json:"device_connected"`

	// Day 1 - Sales Readiness
	ProductsAddedCount int  `gorm:"default:0" json:"products_added_count"`
	PaymentConfigured  bool `gorm:"default:false" json:"payment_configured"`
	ReceiptTested      bool `gorm:"default:false" json:"receipt_tested"`

	// Day 2 - Staff & Control
	CashierCreated bool `gorm:"default:false" json:"cashier_created"`
	LoginTested    bool `gorm:"default:false" json:"login_tested"`

	// Day 3 - Proof of Value
	FirstSaleRecorded bool `gorm:"default:false" json:"first_sale_recorded"`
	ReportViewed      bool `gorm:"default:false" json:"report_viewed"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
