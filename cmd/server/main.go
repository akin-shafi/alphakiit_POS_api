package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"pos-fiber-app/internal/auth"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/config"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/middleware"
	"pos-fiber-app/internal/outlet"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/terminal"
	"pos-fiber-app/internal/user"
	"pos-fiber-app/pkg/database"

	_ "pos-fiber-app/docs"
)

// @title POS System API
// @version 1.0
// @description Multi-tenant POS backend built with Go Fiber
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	db := database.ConnectDB()
	database.RunMigrations(db)

	// Initialize Fiber app
	app := fiber.New()

	// Swagger UI
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// ======================
	// Public routes (no auth required)
	// ======================
	public := app.Group("/api/v1")

	// User routes (e.g., register, profile - if any public ones)
	user.RegisterUserRoutes(public, db)

	// Auth public routes: login, forgot-password, verify-otp, reset-password, refresh
	// Note: RegisterAuthRoutes handles both public and protected parts

	// ======================
	// Protected routes (JWT + Tenant required)
	// ======================
	protected := app.Group("/api/v1",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
	)

	// Register ALL auth routes (public + protected like logout)
	auth.RegisterAuthRoutes(public, protected, db)

	// Migrate password reset OTP table
	auth.MigratePasswordReset(db)

	// Business management (OWNER/MANAGER level)
	business.RegisterBusinessRoutes(protected, db)

	// Outlet & Terminal management
	outlet.RegisterRoutes(protected, db)
	terminal.RegisterRoutes(protected, db)

	// Example protected endpoints
	protected.Get("/reports",
		middleware.RequireRoles("OWNER", "MANAGER"),
		reportsHandler,
	)

	protected.Post("/sales",
		middleware.RequireRoles("CASHIER"),
		createSaleHandler,
	)

	// ======================
	// Business-scoped routes (JWT + Tenant + Current Business required)
	// These are for Categories, Products, Inventory
	// ======================
	businessScoped := app.Group("/api/v1",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
		middleware.CurrentBusinessMiddleware(), // Must come after TenantMiddleware
	)

	// Register business-scoped modules
	category.RegisterCategoryRoutes(businessScoped, db)
	product.RegisterProductRoutes(businessScoped, db)
	inventory.RegisterInventoryRoutes(businessScoped, db)
	sale.RegisterSaleRoutes(businessScoped, db) // ← NEW: Sales module

	// Start server
	log.Println("Server running on port", config.AppPort())
	log.Fatal(app.Listen(":" + config.AppPort()))
}

// ======================
// Placeholder handlers
// ======================

func reportsHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Reports endpoint - accessible to OWNER and MANAGER",
	})
}

func createSaleHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Create sale endpoint - accessible to CASHIER",
	})
}
