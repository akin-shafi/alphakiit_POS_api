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
	"pos-fiber-app/internal/outlet"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/sale"
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

func RunMigrations(db *gorm.DB) {
	err := db.AutoMigrate(
		&auth.RefreshToken{}, // if you have password_reset_otp table
		&user.User{},
		&business.Business{},
		&business.Tenant{}, // if you're still using the Tenant table
		&outlet.Outlet{},
		&terminal.Terminal{},
		&category.Category{},
		&product.Product{},
		&inventory.Inventory{},
		&sale.Sale{},
        &sale.SaleItem{},
        
		// Add any other models here (e.g. password reset OTP if separate)
	)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
}
