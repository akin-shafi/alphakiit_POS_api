package common

type PlanType string

const (
	PlanTrial            PlanType = "TRIAL"
	
	PlanEssentialMonthly PlanType = "ESSENTIAL_MONTHLY"
	PlanEssentialAnnual  PlanType = "ESSENTIAL_ANNUAL"
	
	PlanGrowthMonthly    PlanType = "GROWTH_MONTHLY"
	PlanGrowthAnnual     PlanType = "GROWTH_ANNUAL"
	
	PlanScaleMonthly     PlanType = "SCALE_MONTHLY"
	PlanScaleAnnual      PlanType = "SCALE_ANNUAL"

	// Legacy support / Internal mapping
	PlanMonthly          PlanType = "MONTHLY"
	PlanAnnual           PlanType = "ANNUAL"
	PlanQuarterly        PlanType = "QUARTERLY"
	PlanServiceMonthly   PlanType = "SERVICE_MONTHLY"
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
	ModuleBulkStock  ModuleType = "BULK_STOCK_MANAGEMENT"
)

type SubscriptionPlan struct {
	Type                 PlanType       `json:"type"`
	Name                 string         `json:"name"`
	DurationDays         int            `json:"duration_days"`
	Price                float64        `json:"price"`
	Currency             string         `json:"currency"`
	UserLimit            int            `json:"user_limit"`
	ProductLimit         int            `json:"product_limit"`
	AllowedBusinessTypes []BusinessType `json:"allowed_business_types,omitempty"`
	FreeModules          []ModuleType   `json:"free_modules,omitempty"`
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
	// ESSENTIAL TIER
	{
		Type:         PlanEssentialMonthly,
		Name:         "Essential Monthly",
		DurationDays: 30,
		Price:        12500,
		Currency:     "NGN",
		UserLimit:    3,
		ProductLimit: 500,
		FreeModules:  []ModuleType{ModuleWhatsApp},
	},
	{
		Type:         PlanEssentialAnnual,
		Name:         "Essential Annual",
		DurationDays: 365,
		Price:        100000,
		Currency:     "NGN",
		UserLimit:    3,
		ProductLimit: 500,
		FreeModules:  []ModuleType{ModuleWhatsApp},
	},
	// GROWTH TIER
	{
		Type:         PlanGrowthMonthly,
		Name:         "Growth Monthly",
		DurationDays: 30,
		Price:        25000,
		Currency:     "NGN",
		UserLimit:    10,
		ProductLimit: 2500,
		FreeModules:  []ModuleType{ModuleInventory, ModuleCompliance, ModuleWhatsApp},
	},
	{
		Type:         PlanGrowthAnnual,
		Name:         "Growth Annual",
		DurationDays: 365,
		Price:        200000,
		Currency:     "NGN",
		UserLimit:    10,
		ProductLimit: 2500,
		FreeModules:  []ModuleType{ModuleInventory, ModuleCompliance, ModuleWhatsApp},
	},
	// SCALE TIER
	{
		Type:         PlanScaleMonthly,
		Name:         "Scale Monthly",
		DurationDays: 30,
		Price:        45000,
		Currency:     "NGN",
		UserLimit:    50,
		ProductLimit: 15000,
		FreeModules: []ModuleType{
			ModuleKDS, ModuleTables, ModuleDrafts, ModuleInventory,
			ModuleRecipe, ModuleWhatsApp, ModuleCompliance, ModuleQRMenu, ModuleBulkStock,
		},
	},
	{
		Type:         PlanScaleAnnual,
		Name:         "Scale Annual",
		DurationDays: 365,
		Price:        360000,
		Currency:     "NGN",
		UserLimit:    50,
		ProductLimit: 15000,
		FreeModules: []ModuleType{
			ModuleKDS, ModuleTables, ModuleDrafts, ModuleInventory,
			ModuleRecipe, ModuleWhatsApp, ModuleCompliance, ModuleQRMenu, ModuleBulkStock,
		},
	},
}

type ModulePlan struct {
	Type        ModuleType   `json:"type"`
	Name        string       `json:"name"`
	Price       float64      `json:"price"` // price per month
	Description string       `json:"description"`
	DependsOn   []ModuleType `json:"depends_on,omitempty"`
}

var AvailableModules = []ModulePlan{
	{
		Type:        ModuleKDS,
		Name:        "Kitchen Display System (KDS)",
		Price:       3000,
		Description: "Real-time kitchen order monitor for chefs",
	},
	{
		Type:        ModuleTables,
		Name:        "Table Management",
		Price:       3000,
		Description: "Track floor layouts and table status",
	},
	{
		Type:        ModuleDrafts,
		Name:        "Save Drafts",
		Price:       2500,
		Description: "Save and resume incomplete orders",
	},
	{
		Type:        ModuleInventory,
		Name:        "Advanced Inventory Control",
		Price:       15000,
		Description: "Batch tracking, shrinkage alerts, and stock history",
	},
	{
		Type:        ModuleRecipe,
		Name:        "Recipe & Cost Control (BOM)",
		Price:       12000,
		Description: "Ingredient-level cost tracking per item sold",
		DependsOn:   []ModuleType{ModuleInventory},
	},
	{
		Type:        ModuleWhatsApp,
		Name:        "Security & Owner WhatsApp Alerts",
		Price:       5000,
		Description: "Instant alerts for voids, overrides, refunds, and logins",
	},
	{
		Type:        ModuleCompliance,
		Name:        "Automated Compliance & Audit Replay",
		Price:       12000,
		Description: "Tax-ready reports, audit trail, and activity playback",
	},
	{
		Type:        ModuleQRMenu,
		Name:        "QR Digital Menu",
		Price:       5000,
		Description: "Public QR-based digital menu with live product updates",
	},
	{
		Type:        ModuleBulkStock,
		Name:        "Bulk Stock & Round Tracking",
		Price:       10000,
		Description: "Specialized tracking for fuel, gas, and bulk commodities.",
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
