package subscription

import (
	"pos-fiber-app/internal/common"
	"time"

	"gorm.io/gorm"
)

type PlanType = common.PlanType
type SubscriptionStatus = common.SubscriptionStatus
type ModuleType = common.ModuleType

const (
	PlanTrial            = common.PlanTrial
	PlanMonthly          = common.PlanMonthly
	PlanQuarterly        = common.PlanQuarterly
	PlanAnnual           = common.PlanAnnual
	PlanServiceMonthly   = common.PlanServiceMonthly
	PlanServiceQuarterly = common.PlanServiceQuarterly
	PlanServiceAnnual    = common.PlanServiceAnnual

	StatusActive         = common.StatusActive
	StatusExpired        = common.StatusExpired
	StatusCancelled      = common.StatusCancelled
	StatusGracePeriod    = common.StatusGracePeriod
	StatusPendingPayment = common.StatusPendingPayment

	ModuleKDS        = common.ModuleKDS
	ModuleTables     = common.ModuleTables
	ModuleDrafts     = common.ModuleDrafts
	ModuleInventory  = common.ModuleInventory
	ModuleRecipe     = common.ModuleRecipe
	ModuleWhatsApp   = common.ModuleWhatsApp
	ModuleCompliance = common.ModuleCompliance
	ModuleQRMenu     = common.ModuleQRMenu
)

type BusinessModule struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	BusinessID uint       `gorm:"index;uniqueIndex:idx_bus_mod" json:"business_id"`
	Module     ModuleType `gorm:"type:varchar(50);uniqueIndex:idx_bus_mod" json:"module"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Subscription struct {
	ID                   uint               `gorm:"primaryKey" json:"id"`
	BusinessID           uint               `gorm:"index" json:"business_id"`
	PlanType             PlanType           `gorm:"type:varchar(20)" json:"plan_type"`
	Status               SubscriptionStatus `gorm:"type:varchar(20)" json:"status"`
	StartDate            time.Time          `json:"start_date"`
	EndDate              time.Time          `json:"end_date"`
	AutoRenew            bool               `gorm:"default:false" json:"auto_renew"`
	PaymentMethod        string             `json:"payment_method"`
	TransactionReference string             `json:"transaction_reference"`
	AmountPaid           float64            `gorm:"type:decimal(12,2)" json:"amount_paid"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type SubscriptionPlan = common.SubscriptionPlan

var AvailablePlans = common.AvailablePlans

type ModulePlan = common.ModulePlan

var AvailableModules = common.AvailableModules

type ModuleBundle = common.ModuleBundle

var AvailableBundles = common.AvailableBundles

type PromoCode struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Code               string    `gorm:"uniqueIndex;type:varchar(50)" json:"code"`
	DiscountPercentage float64   `json:"discount_percentage"`
	MaxUses            int       `json:"max_uses"`
	UsedCount          int       `gorm:"default:0" json:"used_count"`
	ExpiryDate         time.Time `json:"expiry_date"`
	Active             bool      `gorm:"default:true" json:"active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ReferralCode struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex;type:varchar(50)" json:"code"`
	InstallerID uint      `gorm:"index" json:"installer_id"`
	UsesCount   int       `gorm:"default:0" json:"uses_count"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommissionStatus string

const (
	CommissionPending   CommissionStatus = "PENDING"
	CommissionPaid      CommissionStatus = "PAID"
	CommissionCancelled CommissionStatus = "CANCELLED"
)

type CommissionRecord struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	InstallerID    uint             `gorm:"index" json:"installer_id"`
	BusinessID     uint             `gorm:"index" json:"business_id"`
	SubscriptionID uint             `gorm:"index" json:"subscription_id"`
	Amount         float64          `gorm:"type:decimal(12,2)" json:"amount"`
	Type           string           `json:"type"` // "ONBOARDING" or "RENEWAL"
	Status         CommissionStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	PaidAt         *time.Time       `json:"paid_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type CommissionSetting struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	OnboardingRate          float64   `gorm:"default:20.0" json:"onboarding_rate"` // in percentage
	RenewalRate             float64   `gorm:"default:10.0" json:"renewal_rate"`    // in percentage
	EnableRenewalCommission bool      `gorm:"default:true" json:"enable_renewal_commission"`
	MinRenewalDays          int       `gorm:"default:0" json:"min_renewal_days"`         // 0 means any plan, 365 means only annual
	CommissionDurationDays  int       `gorm:"default:0" json:"commission_duration_days"` // 0 means lifetime
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type TrainingResource struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	URL         string    `gorm:"type:varchar(500);not null" json:"url"`        // Youtube Link or File Link
	Type        string    `gorm:"type:varchar(50);default:'VIDEO'" json:"type"` // "VIDEO" or "PDF"
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PayoutStatus string

const (
	PayoutRequested  PayoutStatus = "REQUESTED"
	PayoutProcessing PayoutStatus = "PROCESSING"
	PayoutCompleted  PayoutStatus = "COMPLETED"
	PayoutRejected   PayoutStatus = "REJECTED"
)

type PayoutRequest struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	InstallerID uint         `gorm:"index" json:"installer_id"`
	Amount      float64      `gorm:"type:decimal(12,2)" json:"amount"`
	Status      PayoutStatus `gorm:"type:varchar(20);default:'REQUESTED'" json:"status"`
	BankName    string       `json:"bank_name"`
	AccountNum  string       `json:"account_number"`
	AccountName string       `json:"account_name"`
	AdminNote   string       `json:"admin_note"`
	ProcessedAt *time.Time   `json:"processed_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Subscription{},
		&PromoCode{},
		&BusinessModule{},
		&ReferralCode{},
		&CommissionRecord{},
		&CommissionSetting{},
		&TrainingResource{},
		&PayoutRequest{},
	)
}
