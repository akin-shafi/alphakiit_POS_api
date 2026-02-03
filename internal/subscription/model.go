package subscription

import (
	"time"
)

type PlanType string

const (
	PlanTrial     PlanType = "TRIAL"
	PlanMonthly   PlanType = "MONTHLY"
	PlanQuarterly PlanType = "QUARTERLY"
	PlanAnnual    PlanType = "ANNUAL"
)

type SubscriptionStatus string

const (
	StatusActive      SubscriptionStatus = "ACTIVE"
	StatusExpired     SubscriptionStatus = "EXPIRED"
	StatusCancelled   SubscriptionStatus = "CANCELLED"
	StatusGracePeriod SubscriptionStatus = "GRACE_PERIOD"
)

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

type SubscriptionPlan struct {
	Type         PlanType `json:"type"`
	Name         string   `json:"name"`
	DurationDays int      `json:"duration_days"`
	Price        float64  `json:"price"`
	Currency     string   `json:"currency"`
	UserLimit    int      `json:"user_limit"`
	ProductLimit int      `json:"product_limit"`
}

var AvailablePlans = []SubscriptionPlan{
	{
		Type:         PlanTrial,
		Name:         "Free Trial",
		DurationDays: 14,
		Price:        0,
		Currency:     "NGN",
		UserLimit:    2,
		ProductLimit: 50,
	},
	{
		Type:         PlanMonthly,
		Name:         "Monthly Pro",
		DurationDays: 30,
		Price:        5000,
		Currency:     "NGN",
		UserLimit:    5,
		ProductLimit: 500,
	},
	{
		Type:         PlanQuarterly,
		Name:         "Quarterly Pro",
		DurationDays: 90,
		Price:        13500,
		Currency:     "NGN",
		UserLimit:    10,
		ProductLimit: 1000,
	},
	{
		Type:         PlanAnnual,
		Name:         "Annual Pro",
		DurationDays: 365,
		Price:        48000,
		Currency:     "NGN",
		UserLimit:    20,
		ProductLimit: 5000,
	},
}
