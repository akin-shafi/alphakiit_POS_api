// internal/business/seed_data.go
package business

import (
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/product"
)

type SeedTemplate struct {
	Categories []category.Category
	Products   []product.Product
	Inventory  []inventory.Inventory
}

// Helper: generate default inventory entries (one per product)
func generateDefaultInventory(count int, stock int) []inventory.Inventory {
	invs := make([]inventory.Inventory, count)
	for i := range invs {
		invs[i] = inventory.Inventory{
			CurrentStock:  stock,
			LowStockAlert: 10,
		}
	}
	return invs
}

var SeedTemplates = map[common.BusinessType]SeedTemplate{
	common.TypeRestaurant: {
		Categories: []category.Category{
			{Name: "Main Courses", Description: "Rice, Pasta, Soups"},
			{Name: "Proteins", Description: "Beef, Chicken, Fish"},
			{Name: "Swallows", Description: "Amala, Eba, Pounded Yam"},
			{Name: "Drinks", Description: "Soft drinks, Water, Juice"},
			{Name: "Sides", Description: "Plantain, Salad"},
		},
		Products: []product.Product{
			{Name: "Jollof Rice", Price: 1500, SKU: "RICE001"},
			{Name: "Fried Rice", Price: 1500, SKU: "RICE002"},
			{Name: "Egusi Soup", Price: 1200, SKU: "SOUP001"},
			{Name: "Ogbono Soup", Price: 1200, SKU: "SOUP002"},
			{Name: "Goat Meat", Price: 800, SKU: "PROT001"},
			{Name: "Beef", Price: 600, SKU: "PROT002"},
			{Name: "Chicken (Full)", Price: 2500, SKU: "PROT003"},
			{Name: "Fried Fish", Price: 1800, SKU: "PROT004"},
			{Name: "Amala", Price: 400, SKU: "SWAL001"},
			{Name: "Eba", Price: 300, SKU: "SWAL002"},
			{Name: "Pounded Yam", Price: 600, SKU: "SWAL003"},
			{Name: "Coca Cola (Can)", Price: 400, SKU: "DRNK001"},
			{Name: "Bottled Water", Price: 200, SKU: "DRNK002"},
			{Name: "Malt", Price: 500, SKU: "DRNK003"},
			{Name: "Dodo (Plantain)", Price: 400, SKU: "SIDE001"},
		},
		Inventory: generateDefaultInventory(15, 30),
	},

	common.TypeBar: {
		Categories: []category.Category{
			{Name: "Beers"},
			{Name: "Spirits"},
			{Name: "Wines"},
			{Name: "Cocktails"},
			{Name: "Snacks"},
		},
		Products: []product.Product{
			{Name: "Star Lager", Price: 600, SKU: "BEER001"},
			{Name: "Guinness Stout", Price: 700, SKU: "BEER002"},
			{Name: "Heineken", Price: 800, SKU: "BEER003"},
			{Name: "Hennessy VSOP", Price: 25000, SKU: "SPIR001"},
			{Name: "Jack Daniels", Price: 18000, SKU: "SPIR002"},
			{Name: "Smirnoff Vodka", Price: 12000, SKU: "SPIR003"},
			{Name: "Red Wine (House)", Price: 5000, SKU: "WINE001"},
			{Name: "White Wine (House)", Price: 5000, SKU: "WINE002"},
			{Name: "Margarita", Price: 3500, SKU: "COCK001"},
			{Name: "Mojito", Price: 3500, SKU: "COCK002"},
			{Name: "Peanuts", Price: 500, SKU: "SNCK001"},
			{Name: "Plantain Chips", Price: 800, SKU: "SNCK002"},
		},
		Inventory: generateDefaultInventory(12, 30),
	},

	common.TypeSupermarket: {
		Categories: []category.Category{
			{Name: "Groceries", Description: "Rice, Oil, Pasta"},
			{Name: "Beverages"},
			{Name: "Household"},
			{Name: "Personal Care"},
			{Name: "Snacks"},
		},
		Products: []product.Product{
			{Name: "Bag of Rice (50kg)", Price: 45000, SKU: "GROC001"},
			{Name: "Vegetable Oil (5L)", Price: 8000, SKU: "GROC002"},
			{Name: "Spaghetti (Pack)", Price: 600, SKU: "GROC003"},
			{Name: "Coca Cola (1.5L)", Price: 800, SKU: "BEV001"},
			{Name: "Bottled Water (Pack)", Price: 1500, SKU: "BEV002"},
			{Name: "Detergent (Large)", Price: 3500, SKU: "HH001"},
			{Name: "Toilet Paper (Pack)", Price: 2000, SKU: "HH002"},
			{Name: "Soap Bar", Price: 300, SKU: "PC001"},
			{Name: "Toothpaste", Price: 800, SKU: "PC002"},
			{Name: "Biscuits", Price: 500, SKU: "SNK001"},
			{Name: "Chips", Price: 400, SKU: "SNK002"},
		},
		Inventory: generateDefaultInventory(11, 30),
	},

	common.TypeFuelStation: {
		Categories: []category.Category{
			{Name: "Fuels"},
			{Name: "Motor Oils"},
			{Name: "Snacks & Drinks"},
			{Name: "Car Accessories"},
		},
		Products: []product.Product{
			{Name: "Premium Motor Spirit (PMS)", Price: 617, SKU: "FUEL001"}, // per liter
			{Name: "Diesel (AGO)", Price: 850, SKU: "FUEL002"},
			{Name: "Engine Oil (4L)", Price: 12000, SKU: "OIL001"},
			{Name: "Brake Fluid", Price: 2500, SKU: "OIL002"},
			{Name: "Bottled Water", Price: 200, SKU: "SD001"},
			{Name: "Soft Drink", Price: 400, SKU: "SD002"},
			{Name: "Chin Chin", Price: 500, SKU: "SD003"},
			{Name: "Air Freshener", Price: 1500, SKU: "ACC001"},
			{Name: "Phone Charger", Price: 3000, SKU: "ACC002"},
		},
		Inventory: generateDefaultInventory(9, 30),
	},

	common.TypeRetail: {
		Categories: []category.Category{
			{Name: "Clothing"},
			{Name: "Shoes"},
			{Name: "Accessories"},
			{Name: "Electronics"},
		},
		Products: []product.Product{
			{Name: "T-Shirt", Price: 5000, SKU: "CLTH001"},
			{Name: "Jeans", Price: 12000, SKU: "CLTH002"},
			{Name: "Sneakers", Price: 25000, SKU: "SHOE001"},
			{Name: "Sandals", Price: 8000, SKU: "SHOE002"},
			{Name: "Wrist Watch", Price: 15000, SKU: "ACC001"},
			{Name: "Sunglasses", Price: 7000, SKU: "ACC002"},
			{Name: "Phone Case", Price: 3000, SKU: "ELEC001"},
			{Name: "Earphones", Price: 5000, SKU: "ELEC002"},
		},
		Inventory: generateDefaultInventory(8, 30),
	},

	common.TypeHotel: {
		Categories: []category.Category{
			{Name: "Food"},
			{Name: "Beverages"},
			{Name: "Room Service"},
			{Name: "Amenities"},
		},
		Products: []product.Product{
			{Name: "Club Sandwich", Price: 4500, SKU: "FOOD001"},
			{Name: "Burger", Price: 6000, SKU: "FOOD002"},
			{Name: "Beer", Price: 1500, SKU: "BEV001"},
			{Name: "Wine (Glass)", Price: 3000, SKU: "BEV002"},
			{Name: "Extra Towel", Price: 1000, SKU: "SVC001"},
			{Name: "Laundry Service", Price: 5000, SKU: "SVC002"},
			{Name: "Toothbrush Kit", Price: 800, SKU: "AMEN001"},
			{Name: "Shampoo", Price: 1200, SKU: "AMEN002"},
		},
		Inventory: generateDefaultInventory(8, 30),
	},

	common.TypePharmacy: {
		Categories: []category.Category{
			{Name: "Pain Relief"},
			{Name: "Antibiotics"},
			{Name: "Vitamins & Supplements"},
			{Name: "First Aid"},
			{Name: "Baby Care"},
			{Name: "Personal Hygiene"},
		},
		Products: []product.Product{
			{Name: "Paracetamol 500mg (Pack)", Price: 800, SKU: "MED001"},
			{Name: "Ibuprofen 400mg", Price: 1200, SKU: "MED002"},
			{Name: "Amoxicillin 500mg (10 caps)", Price: 2500, SKU: "MED003"},
			{Name: "Vitamin C 1000mg", Price: 3500, SKU: "MED004"},
			{Name: "Plaster Strips (Box)", Price: 600, SKU: "MED005"},
			{Name: "Baby Diapers (Pack)", Price: 8000, SKU: "MED006"},
			{Name: "Sanitary Pads", Price: 1500, SKU: "MED007"},
			{Name: "Hand Sanitizer", Price: 1000, SKU: "MED008"},
		},
		Inventory: generateDefaultInventory(8, 30),
	},

	common.TypeClinic: {
		Categories: []category.Category{
			{Name: "Consultation Fees"},
			{Name: "Medications"},
			{Name: "Medical Supplies"},
			{Name: "Lab Tests"},
		},
		Products: []product.Product{
			{Name: "General Consultation", Price: 5000, SKU: "CONS001"},
			{Name: "Follow-up Visit", Price: 3000, SKU: "CONS002"},
			{Name: "Paracetamol Tablets", Price: 500, SKU: "MED001"},
			{Name: "Antimalarial Drug", Price: 2000, SKU: "MED002"},
			{Name: "Bandage", Price: 300, SKU: "SUP001"},
			{Name: "Syringe", Price: 200, SKU: "SUP002"},
			{Name: "Blood Test", Price: 8000, SKU: "TEST001"},
			{Name: "Urine Test", Price: 3000, SKU: "TEST002"},
		},
		Inventory: generateDefaultInventory(8, 30),
	},

	common.TypeLounge: {
		Categories: []category.Category{
			{Name: "Premium Spirits"},
			{Name: "Cocktails"},
			{Name: "Beers & Wines"},
			{Name: "Hookah"},
			{Name: "Light Bites"},
		},
		Products: []product.Product{
			{Name: "Hennessy XO", Price: 80000, SKU: "LUX001"},
			{Name: "Champagne (Bottle)", Price: 50000, SKU: "LUX002"},
			{Name: "Signature Cocktail", Price: 8000, SKU: "COCK001"},
			{Name: "Craft Beer", Price: 2000, SKU: "BEER001"},
			{Name: "Red Wine (Bottle)", Price: 15000, SKU: "WINE001"},
			{Name: "Hookah Session", Price: 10000, SKU: "HOOK001"},
			{Name: "Chicken Wings", Price: 6000, SKU: "BITE001"},
			{Name: "Cheese Platter", Price: 12000, SKU: "BITE002"},
		},
		Inventory: generateDefaultInventory(8, 30),
	},

	common.TypeLPGStation: {
		Categories: []category.Category{
			{Name: "LPG Refill (Bulk)", Description: "Cooking gas dispensed by weight"},
			{Name: "Pre-filled Cylinders", Description: "Standard cylinder swaps"},
			{Name: "Empty Cylinders", Description: "New gas cylinder sales"},
			{Name: "Accessories", Description: "Regulators, hoses, and burners"},
		},
		Products: []product.Product{
			// The Core trick: Sell Gas in 10g units to allow fractional sales
			// If 1kg = ₦1,200, then 10g = ₦12
			// This allows selling 12.5kg as 1250 units
			{Name: "Cooking Gas (per 10g)", Price: 12, SKU: "GAS-UNIT-10G", Description: "Base unit for gas sales. 100 units = 1kg"},

			// Standard Refills (For quick selection)
			// These could be composites or just separate products if inventory is tracked strictly by weight
			{Name: "LPG Refill 3kg", Price: 3600, SKU: "LPG-REF-3KG"},
			{Name: "LPG Refill 5kg", Price: 6000, SKU: "LPG-REF-5KG"},
			{Name: "LPG Refill 6kg", Price: 7200, SKU: "LPG-REF-6KG"},
			{Name: "LPG Refill 12.5kg", Price: 15000, SKU: "LPG-REF-12.5KG"},
			{Name: "LPG Refill 25kg", Price: 30000, SKU: "LPG-REF-25KG"},
			{Name: "LPG Refill 50kg", Price: 60000, SKU: "LPG-REF-50KG"},

			// Empty Cylinders
			{Name: "New Cylinder 3kg", Price: 12000, SKU: "CYL-NEW-3KG"},
			{Name: "New Cylinder 5kg", Price: 15000, SKU: "CYL-NEW-5KG"},
			{Name: "New Cylinder 6kg", Price: 18000, SKU: "CYL-NEW-6KG"},
			{Name: "New Cylinder 12.5kg", Price: 45000, SKU: "CYL-NEW-12.5KG"},
			{Name: "New Cylinder 25kg", Price: 80000, SKU: "CYL-NEW-25KG"},
			{Name: "New Cylinder 50kg", Price: 150000, SKU: "CYL-NEW-50KG"},

			// Accessories
			{Name: "Gas Regulator (Standard)", Price: 5000, SKU: "ACC-REG-STD"},
			{Name: "Gas Hose (per yard)", Price: 1500, SKU: "ACC-HOSE-YD"},
			{Name: "Hose Clips (Pair)", Price: 500, SKU: "ACC-CLIPS"},
			{Name: "Gas Lighter", Price: 1000, SKU: "ACC-LIGHTER"},
		},
		Inventory: generateDefaultInventory(17, 100), // Higher default stock for testing
	},

	common.TypeBoutique: {
		Categories: []category.Category{
			{Name: "Men's Wear"},
			{Name: "Women's Wear"},
			{Name: "Kid's Wear"},
			{Name: "Accessories"},
			{Name: "Perfumes"},
		},
		Products: []product.Product{
			{Name: "Formal Suit (Navy)", Price: 45000, SKU: "BTQ-MN-001"},
			{Name: "White Dress Shirt", Price: 8500, SKU: "BTQ-MN-002"},
			{Name: "Designer Evening Gown", Price: 65000, SKU: "BTQ-WN-001"},
			{Name: "Floral Summer Dress", Price: 12500, SKU: "BTQ-WN-002"},
			{Name: "School Uniform Set", Price: 15000, SKU: "BTQ-KD-001"},
			{Name: "Leather Belt", Price: 5000, SKU: "BTQ-ACC-001"},
			{Name: "Silk Scarf", Price: 4500, SKU: "BTQ-ACC-002"},
			{Name: "Royal Oud Perfume", Price: 25000, SKU: "BTQ-PRF-001"},
			{Name: "Ocean Breeze Mist", Price: 12000, SKU: "BTQ-PRF-002"},
		},
		Inventory: generateDefaultInventory(9, 20),
	},

	common.TypeBakery: {
		Categories: []category.Category{
			{Name: "Bread"},
			{Name: "Cakes"},
			{Name: "Pastries"},
			{Name: "Cookies"},
			{Name: "Flour & Supplies"},
		},
		Products: []product.Product{
			{Name: "Family Loaf (600g)", Price: 1200, SKU: "BKY-BRD-01"},
			{Name: "Wheat Bread", Price: 1500, SKU: "BKY-BRD-02"},
			{Name: "Chocolate Fudge Cake", Price: 18000, SKU: "BKY-CKE-01"},
			{Name: "Red Velvet Slite", Price: 2500, SKU: "BKY-CKE-02"},
			{Name: "Sausage Roll", Price: 1200, SKU: "BKY-PAS-01"},
			{Name: "Meat Pie", Price: 1500, SKU: "BKY-PAS-02"},
			{Name: "Shortbread (Tin)", Price: 3500, SKU: "BKY-COK-01"},
			{Name: "Ginger Snaps", Price: 800, SKU: "BKY-COK-02"},
			{Name: "Baking Flour (5kg)", Price: 6500, SKU: "BKY-SPL-01"},
		},
		Inventory: generateDefaultInventory(9, 40),
	},

	common.TypeOther: {
		Categories: []category.Category{
			{Name: "General Products"},
			{Name: "Services"},
		},
		Products: []product.Product{
			{Name: "Miscellaneous Item", Price: 1000, SKU: "OTR-ITM-001"},
			{Name: "Service Charge", Price: 5000, SKU: "OTR-SVC-001"},
		},
		Inventory: generateDefaultInventory(2, 100),
	},
}
