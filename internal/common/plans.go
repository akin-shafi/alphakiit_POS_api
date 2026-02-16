package common

type PlanType string

const (
	PlanTrial            PlanType = "TRIAL"
	PlanMonthly          PlanType = "MONTHLY"
	PlanQuarterly        PlanType = "QUARTERLY"
	PlanAnnual           PlanType = "ANNUAL"
	PlanServiceMonthly   PlanType = "SERVICE_MONTHLY"
	PlanServiceQuarterly PlanType = "SERVICE_QUARTERLY"
	PlanServiceAnnual    PlanType = "SERVICE_ANNUAL"
)

type SubscriptionStatus string

const (
	StatusActive         SubscriptionStatus = "ACTIVE"
	StatusExpired        SubscriptionStatus = "EXPIRED"
	StatusCancelled      SubscriptionStatus = "CANCELLED"
	StatusGracePeriod    SubscriptionStatus = "GRACE_PERIOD"
	StatusPendingPayment SubscriptionStatus = "PENDING_PAYMENT"
)

type ModuleType string

const (
	ModuleKDS        ModuleType = "KITCHEN_DISPLAY"
	ModuleTables     ModuleType = "TABLE_MANAGEMENT"
	ModuleDrafts     ModuleType = "SAVE_DRAFTS"
	ModuleInventory  ModuleType = "ADVANCED_INVENTORY"
	ModuleRecipe     ModuleType = "RECIPE_MANAGEMENT"
	ModuleWhatsApp   ModuleType = "WHATSAPP_ALERTS"
	ModuleCompliance ModuleType = "AUTOMATED_COMPLIANCE"
	ModuleQRMenu     ModuleType = "DIGITAL_MENU_QR"
)

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
		Name:         "Monthly Starter",
		DurationDays: 30,
		Price:        30000,
		Currency:     "NGN",
		UserLimit:    3,
		ProductLimit: 300,
	},
	{
		Type:         PlanQuarterly,
		Name:         "Quarterly Control",
		DurationDays: 90,
		Price:        81000, // ₦27,000 / month
		Currency:     "NGN",
		UserLimit:    7,
		ProductLimit: 1500,
	},
	{
		Type:         PlanAnnual,
		Name:         "Annual Owner Pro",
		DurationDays: 365,
		Price:        300000, // ₦25,000 / month
		Currency:     "NGN",
		UserLimit:    15,
		ProductLimit: 5000,
	},
	{
		Type:         PlanServiceMonthly,
		Name:         "Basic Sales POS (Monthly)",
		DurationDays: 30,
		Price:        15000,
		Currency:     "NGN",
		UserLimit:    2,
		ProductLimit: 25, // Strictly for kiosks/LPG/small shops
	},
	{
		Type:         PlanServiceQuarterly,
		Name:         "Basic Sales POS (Quarterly)",
		DurationDays: 90,
		Price:        40000,
		Currency:     "NGN",
		UserLimit:    3, // +1 User Bonus
		ProductLimit: 25,
	},
	{
		Type:         PlanServiceAnnual,
		Name:         "Basic Sales POS (Annual)",
		DurationDays: 365,
		Price:        150000,
		Currency:     "NGN",
		UserLimit:    5, // +3 User Bonus
		ProductLimit: 25,
	},
}

type ModulePlan struct {
	Type        ModuleType `json:"type"`
	Name        string     `json:"name"`
	Price       float64    `json:"price"` // price per month
	Description string     `json:"description"`
}

var AvailableModules = []ModulePlan{
	{
		Type:        ModuleKDS,
		Name:        "Kitchen Display System (KDS)",
		Price:       5000,
		Description: "Real-time kitchen order monitor for chefs",
	},
	{
		Type:        ModuleTables,
		Name:        "Table Management",
		Price:       5000,
		Description: "Track floor layouts and table status",
	},
	{
		Type:        ModuleDrafts,
		Name:        "Save Drafts",
		Price:       4000,
		Description: "Save and resume incomplete orders",
	},
	{
		Type:        ModuleInventory,
		Name:        "Advanced Inventory Control",
		Price:       18000,
		Description: "Batch tracking, shrinkage alerts, and stock history",
	},
	{
		Type:        ModuleRecipe,
		Name:        "Recipe & Cost Control (BOM)",
		Price:       15000,
		Description: "Ingredient-level cost tracking per item sold",
	},
	{
		Type:        ModuleWhatsApp,
		Name:        "Security & Owner WhatsApp Alerts",
		Price:       8000,
		Description: "Instant alerts for voids, overrides, refunds, and logins",
	},
	{
		Type:        ModuleCompliance,
		Name:        "Automated Compliance & Audit Replay",
		Price:       15000,
		Description: "Tax-ready reports, audit trail, and activity playback",
	},
	{
		Type:        ModuleQRMenu,
		Name:        "QR Digital Menu",
		Price:       8000,
		Description: "Public QR-based digital menu with live product updates",
	},
}

type ModuleBundle struct {
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Price       float64      `json:"price"` // monthly
	Modules     []ModuleType `json:"modules"`
	Description string       `json:"description"`
}

var AvailableBundles = []ModuleBundle{
	{
		Code:  "OPS_PACK",
		Name:  "Operations Pack",
		Price: 12000,
		Modules: []ModuleType{
			ModuleKDS,
			ModuleDrafts,
			ModuleTables,
		},
		Description: "Faster service and smoother order handling",
	},
	{
		Code:  "CONTROL_PACK",
		Name:  "Control & Anti-Loss Pack",
		Price: 30000,
		Modules: []ModuleType{
			ModuleInventory,
			ModuleRecipe,
			ModuleWhatsApp,
		},
		Description: "Prevent stock loss and staff fraud",
	},
	{
		Code:  "COMPLIANCE_PACK",
		Name:  "Compliance & Finance Pack",
		Price: 25000,
		Modules: []ModuleType{
			ModuleInventory,
			ModuleCompliance,
		},
		Description: "Audit-ready reports and compliance",
	},
	{
		Code:  "OWNER_PRO",
		Name:  "Owner Pro Pack",
		Price: 45000,
		Modules: []ModuleType{
			ModuleInventory,
			ModuleRecipe,
			ModuleWhatsApp,
			ModuleCompliance,
			ModuleDrafts,
		},
		Description: "Total business control for owners",
	},
}
