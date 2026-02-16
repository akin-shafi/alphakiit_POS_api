// pkg/database/database.go
package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pos-fiber-app/internal/auth" // if you have password_reset_otp table
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/config"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/outlet"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/recipe"
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/shift"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/table"
	"pos-fiber-app/internal/terminal"
	"pos-fiber-app/internal/tutorial"
	"pos-fiber-app/internal/user"
)

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.DatabaseURL()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	return db
}

func RunMigrations(db *gorm.DB) error {
	// Order is important: migrate parent tables first
	err := db.AutoMigrate(
		&otp.OTP{},
		&auth.RefreshToken{}, // if you have password_reset_otp table
		&user.User{},
		&business.Business{},
		&business.Tenant{}, // if you're still using the Tenant table
		&outlet.Outlet{},
		&terminal.Terminal{},
		&terminal.Printer{},
		&category.Category{},
		&product.Product{},
		&inventory.Inventory{},
		&inventory.StockReservation{}, // NEW: Stock reservations
		&sale.Sale{},
		&sale.SaleItem{},
		&sale.SaleSummary{},        // NEW: Sale summaries for archiving
		&sale.SaleActivityLog{},    // NEW: Sale activity logs
		&shift.Shift{},             // NEW: Shift management
		&shift.ShiftReading{},      // NEW: Shift readings for fuel/gas stations
		&table.Table{},             // NEW: Table management
		&recipe.RecipeIngredient{}, // NEW: Recipe management (BOM)
		&subscription.Subscription{},
		&subscription.PromoCode{},
		&subscription.BusinessModule{},
		&subscription.ReferralCode{},
		&subscription.CommissionRecord{},
		&subscription.CommissionSetting{},
		&subscription.TrainingResource{},
		&subscription.PayoutRequest{},
		&tutorial.Tutorial{},
	)

	if err != nil {
		return err
	}

	// Fallsafe: Manually ensure outlet_id exists in sales table if AutoMigrate skipped it
	if !db.Migrator().HasColumn(&sale.Sale{}, "OutletID") {
		log.Println("Migrator: adding missing outlet_id column to sales table")
		if err := db.Migrator().AddColumn(&sale.Sale{}, "OutletID"); err != nil {
			log.Printf("Warning: Failed to add outlet_id column: %v", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
