package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"pos-fiber-app/docs"
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

	_ "pos-fiber-app/docs" // Imports generated swagger docs
)

// @title POS System API
// @version 1.0
// @description Multi-tenant POS backend built with Go Fiber
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	db := database.ConnectDB()
	// database.RunMigrations(db)

	// Initialize Fiber app
	app := fiber.New()

	// Swagger UI endpoint
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// === DYNAMIC SWAGGER HOST ===
	// Override the hardcoded @host with runtime environment value
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "http://localhost:" + config.AppPort() // fallback for local development
	}

	// Modify generated swagger info at runtime
	docs.SwaggerInfo.Host = apiHost
	docs.SwaggerInfo.BasePath = "/api/v1"

	// ======================
	// Public routes (no authentication required)
	// ======================
	public := app.Group("/api/v1")

	// User-related public routes (if any, e.g., registration)
	user.RegisterUserRoutes(public, db)

	// ======================
	// Protected: Tenant-level resources
	// Requires: JWT + Tenant extraction
	// ======================
	protected := app.Group("/api/v1",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
	)

	// Register auth routes (handles both public login endpoints and protected logout)
	auth.RegisterAuthRoutes(public, protected, db)

	// Migrate password reset table
	auth.MigratePasswordReset(db)

	// Tenant-level resources
	business.RegisterBusinessRoutes(protected, db)
	outlet.RegisterRoutes(protected, db)
	terminal.RegisterRoutes(protected, db)

	// ======================
	// Business-scoped routes
	// Requires: JWT + Tenant + Current Business (via X-Current-Business-ID header)
	// ======================
	businessScoped := app.Group("/api/v1",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
		middleware.CurrentBusinessMiddleware(),
	)

	// Business-specific modules
	category.RegisterCategoryRoutes(businessScoped, db)
	product.RegisterProductRoutes(businessScoped, db)
	inventory.RegisterInventoryRoutes(businessScoped, db)
	sale.RegisterSaleRoutes(businessScoped, db)

	// Start server
	log.Printf("Server starting on %s", apiHost)
	log.Printf("Swagger UI available at %s/swagger/index.html", apiHost)
	log.Fatal(app.Listen(":" + config.AppPort()))
}

// ======================
// Placeholder handlers
// ======================

// func reportsHandler(c *fiber.Ctx) error {
// 	return c.JSON(fiber.Map{
// 		"message": "Reports endpoint - accessible to OWNER and MANAGER",
// 	})
// }

// func createSaleHandler(c *fiber.Ctx) error {
// 	return c.JSON(fiber.Map{
// 		"message": "Create sale endpoint - accessible to CASHIER",
// 	})
// }
