package tutorial

import (
	"pos-fiber-app/internal/common"

	"gorm.io/gorm"
)

func SeedTutorials(db *gorm.DB) error {
	var count int64
	db.Model(&Tutorial{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	tutorials := []Tutorial{
		// LPG STATION
		{
			BusinessType: common.TypeLPGStation,
			Topic:        "Sales",
			Title:        "Selling Gas by Amount or Weight",
			Content:      "When selling gas, the system is flexible. You can enter the amount in Naira (e.g., ₦5,000) or weight in Kilograms (e.g., 12.5kg). The system converts everything to 10g base units for accurate stock tracking.\n\nExample: If price is ₦1,200/kg (₦12 per unit), entering ₦5,000 will add 416 units to the cart.",
			DisplayOrder: 1,
		},
		{
			BusinessType: common.TypeLPGStation,
			Topic:        "Stock",
			Title:        "Meter Readings",
			Content:      "At the end of every shift, the system will prompt for the closing meter reading of the main pump. This allows the business owner to reconcile physical gas depletion with recorded sales.\n\nExample: Opening reading 10,000kg, Closing reading 10,500kg. Sales should account for 500kg depletion.",
			DisplayOrder: 2,
		},
		{
			BusinessType: common.TypeLPGStation,
			Topic:        "Flow",
			Title:        "Shift Reconciliation",
			Content:      "Close your shift daily with the exact cash in your drawer. The system compares this with 'Expected Cash' (Starting Cash + Sales) and reports any variance directly to the owner via email and WhatsApp.",
			DisplayOrder: 3,
		},

		// RESTAURANT
		{
			BusinessType: common.TypeRestaurant,
			Topic:        "Sales",
			Title:        "Table Management",
			Content:      "Assign orders to specific tables to keep track of guest bills. You can print orders directly to the kitchen (KDS) and generate individual or group receipts upon checkout.",
			DisplayOrder: 1,
		},
		{
			BusinessType: common.TypeRestaurant,
			Topic:        "Production",
			Title:        "Recipe & BOM",
			Content:      "For items like 'Jollof Rice', you can define a recipe. Every plate sold will automatically deduct the required portions of rice, oil, and spices from your raw material inventory.\n\nExample: 1 plate = 250g Rice + 50ml Veg Oil.",
			DisplayOrder: 2,
		},

		// LOUNGE / BAR
		{
			BusinessType: common.TypeLounge,
			Topic:        "Sales",
			Title:        "Draft Orders",
			Content:      "In a lounge setting, customers often order over time. Use the 'Save to Draft' feature to keep a bill open and add items as they order more drinks or snacks.",
			DisplayOrder: 1,
		},
		{
			BusinessType: common.TypeLounge,
			Topic:        "Control",
			Title:        "Drink Recipes",
			Content:      "Perfect for cocktails! Define ingredients for each drink. Selling one 'Mojito' will deduct the exact portions of Mint, Rum, and Lime from your store.",
			DisplayOrder: 2,
		},

		// BAKERY
		{
			BusinessType: common.TypeBakery,
			Topic:        "Production",
			Title:        "Production Guidelines",
			Content:      "Convert raw materials (Flour, Sugar, Yeast) into finished goods (Bread, Cakes). Use the production module to record batches, ensuring that your finished goods stock increases while your raw materials decrease correctly.",
			DisplayOrder: 1,
		},
	}

	for _, t := range tutorials {
		db.Create(&t)
	}

	return nil
}
