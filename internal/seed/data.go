package seed

import (
	"pos-fiber-app/internal/common"
)

type SampleProduct struct {
	Name      string
	Price     float64
	Cost      float64
	Stock     int
	SKUPrefix string
}

type SampleCategory struct {
	Name     string
	Products []SampleProduct
}

var sampleData = map[common.BusinessType][]SampleCategory{
	common.TypeRestaurant: {
		{
			Name: "Main Course",
			Products: []SampleProduct{
				{Name: "Jollof Rice", Price: 4500, Cost: 2500, Stock: 50, SKUPrefix: "MAI"},
				{Name: "White Rice", Price: 4000, Cost: 2200, Stock: 40, SKUPrefix: "MAI"},
				{Name: "Fried Rice", Price: 4000, Cost: 2200, Stock: 40, SKUPrefix: "MAI"},
				{Name: "Egusi Soup", Price: 5000, Cost: 3000, Stock: 30, SKUPrefix: "MAI"},
				{Name: "Chicken Burger & Fries", Price: 3500, Cost: 1800, Stock: 25, SKUPrefix: "MAI"},
			},
		},
		{
			Name: "Sides & Appetizers",
			Products: []SampleProduct{
				{Name: "Spring Rolls (Set of 4)", Price: 1500, Cost: 800, Stock: 100, SKUPrefix: "SID"},
				{Name: "Samosas (Set of 4)", Price: 1500, Cost: 800, Stock: 100, SKUPrefix: "SID"},
				{Name: "Plantain Dodo", Price: 800, Cost: 300, Stock: 60, SKUPrefix: "SID"},
			},
		},
		{
			Name: "Beverages",
			Products: []SampleProduct{
				{Name: "Fresh Orange Juice", Price: 1200, Cost: 500, Stock: 40, SKUPrefix: "BEV"},
				{Name: "Chapman Cocktail", Price: 2500, Cost: 1000, Stock: 30, SKUPrefix: "BEV"},
				{Name: "Bottled Water 75cl", Price: 300, Cost: 150, Stock: 200, SKUPrefix: "BEV"},
			},
		},
		{
			Name: "Proteins",
			Products: []SampleProduct{
				{Name: "Beef", Price: 1200, Cost: 800, Stock: 50, SKUPrefix: "PRO"},
				{Name: "Chicken", Price: 800, Cost: 400, Stock: 40, SKUPrefix: "PRO"},
				{Name: "Fish", Price: 500, Cost: 250, Stock: 60, SKUPrefix: "PRO"},
			},
		},
		{
			Name: "Sides",
			Products: []SampleProduct{
				{Name: "Plantain", Price: 800, Cost: 300, Stock: 60, SKUPrefix: "SID"},
				{Name: "Salad", Price: 1200, Cost: 500, Stock: 40, SKUPrefix: "SID"},
			},
		},
	},
	common.TypePharmacy: {
		{
			Name: "Pain Relief",
			Products: []SampleProduct{
				{Name: "Panadol Extra (Pack of 12)", Price: 1200, Cost: 800, Stock: 50, SKUPrefix: "PAI"},
				{Name: "Ibuprofen 400mg", Price: 800, Cost: 400, Stock: 40, SKUPrefix: "PAI"},
				{Name: "Felvin 20mg", Price: 500, Cost: 250, Stock: 60, SKUPrefix: "PAI"},
			},
		},
		{
			Name: "Supplements",
			Products: []SampleProduct{
				{Name: "Vitamin C 1000mg", Price: 2500, Cost: 1500, Stock: 30, SKUPrefix: "SUP"},
				{Name: "Multi-Vitamin Complex", Price: 4500, Cost: 3000, Stock: 20, SKUPrefix: "SUP"},
				{Name: "Cod Liver Oil Capsules", Price: 3200, Cost: 2000, Stock: 25, SKUPrefix: "SUP"},
			},
		},
		{
			Name: "Personal Care",
			Products: []SampleProduct{
				{Name: "Antiseptic Liquid 500ml", Price: 1800, Cost: 1200, Stock: 40, SKUPrefix: "PER"},
				{Name: "Hand Sanitizer 100ml", Price: 700, Cost: 350, Stock: 100, SKUPrefix: "PER"},
				{Name: "Face Masks (Box of 50)", Price: 3500, Cost: 2000, Stock: 15, SKUPrefix: "PER"},
				{Name: "Digital Thermometer", Price: 5000, Cost: 3000, Stock: 10, SKUPrefix: "PER"},
			},
		},
	},
	common.TypeRetail: {
		{
			Name: "Laptops & Accessories",
			Products: []SampleProduct{
				{Name: "Wireless Mouse", Price: 7500, Cost: 4000, Stock: 20, SKUPrefix: "LAP"},
				{Name: "Laptop Backpack", Price: 12000, Cost: 7000, Stock: 15, SKUPrefix: "LAP"},
				{Name: "USB-C Hub Multi-port", Price: 15000, Cost: 9000, Stock: 10, SKUPrefix: "LAP"},
			},
		},
		{
			Name: "Mobile Accessories",
			Products: []SampleProduct{
				{Name: "Fast Charger 20W", Price: 5500, Cost: 2500, Stock: 30, SKUPrefix: "MOB"},
				{Name: "Screen Protector (iPhone 13)", Price: 2500, Cost: 800, Stock: 50, SKUPrefix: "MOB"},
				{Name: "Bluetooth Earbuds", Price: 18000, Cost: 10000, Stock: 12, SKUPrefix: "MOB"},
				{Name: "Phone Tripod Stand", Price: 4500, Cost: 2000, Stock: 20, SKUPrefix: "MOB"},
			},
		},
		{
			Name: "Home Essentials",
			Products: []SampleProduct{
				{Name: "LED Table Lamp", Price: 8500, Cost: 4500, Stock: 15, SKUPrefix: "HOM"},
				{Name: "Extension Cord cable", Price: 4000, Cost: 2200, Stock: 25, SKUPrefix: "HOM"},
				{Name: "Digital Wall Clock", Price: 6000, Cost: 3500, Stock: 10, SKUPrefix: "HOM"},
			},
		},
	},
	common.TypeSupermarket: {
		{
			Name: "Groceries",
			Products: []SampleProduct{
				{Name: "Golden Penny Pasta 500g", Price: 850, Cost: 700, Stock: 200, SKUPrefix: "GRO"},
				{Name: "Indomie Instant Noodles (Pack of 10)", Price: 4500, Cost: 3800, Stock: 50, SKUPrefix: "GRO"},
				{Name: "Peak Milk Powder 400g", Price: 3200, Cost: 2800, Stock: 40, SKUPrefix: "GRO"},
				{Name: "Milo Beverage 400g", Price: 2800, Cost: 2400, Stock: 45, SKUPrefix: "GRO"},
			},
		},
		{
			Name: "Household",
			Products: []SampleProduct{
				{Name: "Ariel Detergent 1kg", Price: 2500, Cost: 2000, Stock: 60, SKUPrefix: "HOU"},
				{Name: "Liquid Dish Wash 500ml", Price: 1200, Cost: 800, Stock: 80, SKUPrefix: "HOU"},
				{Name: "Paper Towels (Pack of 2)", Price: 1500, Cost: 1000, Stock: 50, SKUPrefix: "HOU"},
			},
		},
		{
			Name: "Drinks & Snacks",
			Products: []SampleProduct{
				{Name: "Lays Classic Chips", Price: 1800, Cost: 1300, Stock: 40, SKUPrefix: "DRI"},
				{Name: "Coca Cola 1.5L", Price: 800, Cost: 650, Stock: 100, SKUPrefix: "DRI"},
				{Name: "Heineken Beer 33cl", Price: 1200, Cost: 900, Stock: 72, SKUPrefix: "DRI"},
			},
		},
	},
	common.TypeBoutique: {
		{
			Name: "Menswear",
			Products: []SampleProduct{
				{Name: "Cotton Polo Shirt", Price: 8500, Cost: 4500, Stock: 30, SKUPrefix: "MEN"},
				{Name: "Slim Fit Chinos", Price: 15000, Cost: 9000, Stock: 20, SKUPrefix: "MEN"},
				{Name: "Casual Blazer", Price: 35000, Cost: 20000, Stock: 5, SKUPrefix: "MEN"},
				{Name: "Leather Belt", Price: 6500, Cost: 3000, Stock: 25, SKUPrefix: "MEN"},
			},
		},
		{
			Name: "Womenswear",
			Products: []SampleProduct{
				{Name: "Floral Summer Dress", Price: 12500, Cost: 7000, Stock: 15, SKUPrefix: "WOM"},
				{Name: "High Waist Jeans", Price: 18000, Cost: 10000, Stock: 20, SKUPrefix: "WOM"},
				{Name: "Silk Scarf", Price: 4500, Cost: 2000, Stock: 30, SKUPrefix: "WOM"},
			},
		},
		{
			Name: "Accessories",
			Products: []SampleProduct{
				{Name: "Luxury Wristwatch", Price: 55000, Cost: 35000, Stock: 8, SKUPrefix: "ACC"},
				{Name: "Sunglasses (UV400)", Price: 7500, Cost: 3500, Stock: 20, SKUPrefix: "ACC"},
				{Name: "Crossbody Bag", Price: 12000, Cost: 6000, Stock: 12, SKUPrefix: "ACC"},
			},
		},
	},
	common.TypeFuelStation: {
		{
			Name: "Lubricants",
			Products: []SampleProduct{
				{Name: "Engine Oil 5W-30 (4L)", Price: 25000, Cost: 18000, Stock: 20, SKUPrefix: "LUB"},
				{Name: "Brake Fluid 500ml", Price: 3500, Cost: 2000, Stock: 40, SKUPrefix: "LUB"},
				{Name: "Coolant concentrate (1L)", Price: 4500, Cost: 2500, Stock: 30, SKUPrefix: "LUB"},
				{Name: "Gear Oil 80W-90", Price: 6000, Cost: 4000, Stock: 15, SKUPrefix: "LUB"},
			},
		},
		{
			Name: "Station Store",
			Products: []SampleProduct{
				{Name: "Car Air Freshener", Price: 2500, Cost: 1200, Stock: 50, SKUPrefix: "STO"},
				{Name: "Wiper Fluid", Price: 3000, Cost: 1800, Stock: 25, SKUPrefix: "STO"},
				{Name: "Tire Pressure Gauge", Price: 5000, Cost: 3000, Stock: 10, SKUPrefix: "STO"},
			},
		},
		{
			Name: "Travel Snacks",
			Products: []SampleProduct{
				{Name: "Mineral Water (Single)", Price: 200, Cost: 100, Stock: 100, SKUPrefix: "TRA"},
				{Name: "Mixed Nuts Pack", Price: 1500, Cost: 900, Stock: 40, SKUPrefix: "TRA"},
				{Name: "Energy Drink 25cl", Price: 800, Cost: 500, Stock: 60, SKUPrefix: "TRA"},
			},
		},
	},
	common.TypeLPGStation: {
		{
			Name: "Bulk Gas Refills",
			Products: []SampleProduct{
				{Name: "LPG Refill 0.5kg",   Price: 600,   Cost: 400,   Stock: 100, SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 1kg",     Price: 1200,  Cost: 800,   Stock: 100, SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 1.5kg",   Price: 1800,  Cost: 1200,  Stock: 100, SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 3kg",     Price: 3600,  Cost: 2400,  Stock: 50,  SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 5kg",     Price: 6000,  Cost: 4000,  Stock: 30,  SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 6kg",     Price: 7200,  Cost: 4800,  Stock: 20,  SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 12.5kg",  Price: 15000, Cost: 10000, Stock: 8,   SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 25kg",    Price: 30000, Cost: 21000, Stock: 4,   SKUPrefix: "GAS-REF"},
				{Name: "LPG Refill 50kg",    Price: 60000, Cost: 42000, Stock: 2,   SKUPrefix: "GAS-REF"},
			},
		},

		{
			Name: "Gas Accessories",
			Products: []SampleProduct{
				{Name: "Gas Regulator (Standard)", Price: 5000, Cost: 3500, Stock: 15, SKUPrefix: "ACC"},
				{Name: "Gas Hose (per yard)", Price: 1500, Cost: 1000, Stock: 50, SKUPrefix: "ACC"},
				{Name: "Gas Lighter", Price: 1000, Cost: 500, Stock: 40, SKUPrefix: "ACC"},
			},
		},
	},
	common.TypeBakery: {
		{
			Name: "Fresh Bread",
			Products: []SampleProduct{
				{Name: "Family Loaf (600g)", Price: 1200, Cost: 800, Stock: 40, SKUPrefix: "BRD"},
				{Name: "Wheat Bread", Price: 1500, Cost: 1000, Stock: 25, SKUPrefix: "BRD"},
				{Name: "Coconut Bread", Price: 1800, Cost: 1200, Stock: 15, SKUPrefix: "BRD"},
			},
		},
		{
			Name: "Pastries & Cakes",
			Products: []SampleProduct{
				{Name: "Sausage Roll", Price: 1200, Cost: 700, Stock: 50, SKUPrefix: "PAS"},
				{Name: "Meat Pie", Price: 1500, Cost: 900, Stock: 50, SKUPrefix: "PAS"},
				{Name: "Chocolate Fudge Cake", Price: 18000, Cost: 12000, Stock: 5, SKUPrefix: "CKE"},
			},
		},
	},
	common.TypeOther: {
		{
			Name: "General Supplies",
			Products: []SampleProduct{
				{Name: "Standard Notebook A5", Price: 1500, Cost: 800, Stock: 100, SKUPrefix: "GEN"},
				{Name: "Ballpoint Pen (Pack of 10)", Price: 2500, Cost: 1500, Stock: 50, SKUPrefix: "GEN"},
				{Name: "Standard Calculator", Price: 4500, Cost: 2500, Stock: 20, SKUPrefix: "GEN"},
				{Name: "Desk Organizer", Price: 6500, Cost: 3500, Stock: 15, SKUPrefix: "GEN"},
			},
		},
		{
			Name: "Misc Utilities",
			Products: []SampleProduct{
				{Name: "Rechargeable Torch", Price: 8500, Cost: 4500, Stock: 20, SKUPrefix: "MIS"},
				{Name: "Multi-tool Kit", Price: 12000, Cost: 7000, Stock: 10, SKUPrefix: "MIS"},
				{Name: "Portable Padlock", Price: 3500, Cost: 1500, Stock: 30, SKUPrefix: "MIS"},
			},
		},
		{
			Name: "Cleaning Tools",
			Products: []SampleProduct{
				{Name: "Microfiber Cloth Set", Price: 2500, Cost: 1200, Stock: 40, SKUPrefix: "CLE"},
				{Name: "Spray Bottle 500ml", Price: 1200, Cost: 600, Stock: 60, SKUPrefix: "CLE"},
				{Name: "Compact Dustpan & Brush", Price: 3500, Cost: 1800, Stock: 15, SKUPrefix: "CLE"},
			},
		},
	},
}
