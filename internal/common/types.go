// internal/common/types.go
package common

type BusinessType string
type Currency string

// BusinessType constants - used across the app (business, seed, validation, etc.)
const (
	TypeRestaurant  BusinessType = "RESTAURANT"
	TypeBar         BusinessType = "BAR"
	TypeSupermarket BusinessType = "SUPERMARKET"
	TypeLounge      BusinessType = "LOUNGE"
	TypeFuelStation BusinessType = "FUEL_STATION"
	TypeRetail      BusinessType = "RETAIL"
	TypeHotel       BusinessType = "HOTEL"
	TypePharmacy    BusinessType = "PHARMACY"
	TypeClinic      BusinessType = "CLINIC"
	TypeLPGStation  BusinessType = "LPG_STATION"
	TypeBoutique    BusinessType = "BOUTIQUE"
	TypeBakery      BusinessType = "BAKERY"
	TypeOther       BusinessType = "OTHER"
	// Add more business types here in the future
)

// Currency constants - used for pricing and display
const (
	CurrencyNGN Currency = "NGN" // Nigerian Naira
	CurrencyUSD Currency = "USD" // US Dollar
	CurrencyGBP Currency = "GBP" // British Pound
	CurrencyEUR Currency = "EUR" // Euro
	// Add more currencies as needed
)
