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
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/shift"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/table"
	"pos-fiber-app/internal/terminal"
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
		&category.Category{},
		&product.Product{},
		&inventory.Inventory{},
		&inventory.StockReservation{}, // NEW: Stock reservations
		&sale.Sale{},
		&sale.SaleItem{},
		&sale.SaleActivityLog{}, // NEW: Sale activity logs
		&shift.Shift{},          // NEW: Shift management
		&table.Table{},          // NEW: Table management
		&subscription.Subscription{},

		// add all other models here...
		// &models.Transaction{},
		// &models.Payment{},
		// &models.Customer{},
		// etc.
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
