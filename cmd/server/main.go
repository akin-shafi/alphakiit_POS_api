package main

import (
	"log"
	"os"
	"strings"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	// NEW: Modern Swaggo for Fiber v2/v3 + OpenAPI 3

	"pos-fiber-app/docs" // Keep this for swag init to generate docs
	"pos-fiber-app/internal/advert"
	"pos-fiber-app/internal/archiver"
	"pos-fiber-app/internal/auth"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/config"
	"pos-fiber-app/internal/expense"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/middleware"
	"pos-fiber-app/internal/notification"
	"pos-fiber-app/internal/onboarding"
	"pos-fiber-app/internal/outlet"
	"pos-fiber-app/internal/printing"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/recipe"
	"pos-fiber-app/internal/report"
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/seed"
	"pos-fiber-app/internal/shift" // NEW: Shift management
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/table" // NEW: Table management
	"pos-fiber-app/internal/terminal"
	"pos-fiber-app/internal/tutorial"
	"pos-fiber-app/internal/user"
	"pos-fiber-app/pkg/database"

	_ "pos-fiber-app/docs" // Required for swag to include docs
)

// ==================== Swagger / OpenAPI 3 Annotations ====================

// @title                       POS System API
// @version                     1.0
// @description                 Multi-tenant POS backend built with Fiber
// @termsOfService              http://swagger.io/terms/

// @contact.name                API Support
// @contact.url                 http://www.swagger.io/support
// @contact.email               support@swagger.io

// @license.name                MIT
// @license.url                 https://opensource.org/licenses/MIT

// @host                        localhost:5050
// @BasePath                    /api/v1
// @schemes                     http https

// Proper Bearer JWT Authentication (OpenAPI 3)
// @securityDefinitions.bearer  BearerAuth
// @in                          header
// @name                        Authorization

// @securityDefinitions.apiKey BusinessID
// @in                          header
// @name                        X-Current-Business-ID
// @description                 JWT Authorization header using the Bearer scheme. Enter your token only (without "Bearer " prefix).

func main() {
	config.LoadEnv()

	db := database.ConnectDB()
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	// tutorial.Migrate(db)       // Ensure tutorials table exists
	// tutorial.SeedTutorials(db) // Seed tutorial content (after migrations)
	// subscription.Migrate(db)   // Ensure subscription/training resources tables exist
	// if err := seed.SeedInstallerData(db); err != nil {
	// 	log.Printf("[SEED ERROR] Failed to seed installer data: %v", err)
	// } else {
	// 	log.Println("[SEED] Installer data seeded successfully")
	// }

	// === Start Background Tasks ===
	archiver.StartDataLifecycleManager(db)
	// report.StartReportScheduler(db) // DEPRECATED: Now handled by external scheduler service via API

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Return status code with error message in JSON
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// === Global Recover (Prevents crashes) ===
	app.Use(recover.New())

	// === Global Rate Limiter ===
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later.",
			})
		},
	}))

	// === CORS ===

	app.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Current-Business-ID, X-Current-Business-Id, X-Business-ID, X-Business-Id, X-Tenant-ID, X-Tenant-Id",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, X-Current-Business-ID, X-Current-Business-Id, X-Business-ID, X-Business-Id, X-Tenant-ID, X-Tenant-Id",
	}))

	// === NEW: Modern Swagger UI with OpenAPI 3 support ===
	app.Get("/swagger/*", swagger.HandlerDefault) // UI at /swagger/index.html// Serves at /swagger/index.html

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "POS API running",
		})
	})

	// === Dynamic Swagger config (host & scheme) ===
	setDynamicSwaggerConfig()

	// === Routes ===
	apiV1 := app.Group("/api/v1")

	// Internal Trigger Endpoints (for external scheduler)
	reportService := report.NewReportService(db)
	reportCtrl := report.NewInternalReportController(db, reportService)
	internal := apiV1.Group("/internal", reportCtrl.AuthMiddleware)
	internal.Post("/cron/daily-report", reportCtrl.DailyReportHandler)
	internal.Post("/cron/weekly-audit-reminder", reportCtrl.WeeklyAuditHandler)
	internal.Post("/cron/monthly-report-reminder", reportCtrl.MonthlyReportHandler)

	// 1. PUBLIC ROUTES (No Auth Required)
	// --------------------------------------------------
	business.RegisterPublicBusinessRoutes(apiV1, db)
	onboarding.RegisterRoutes(apiV1, db)
	subscription.RegisterPublicRoutes(apiV1, db)
	seed.RegisterPublicRoutes(apiV1, db)
	advert.RegisterPublicAdvertRoutes(apiV1, db)
	product.RegisterPublicProductRoutes(apiV1, db)

	// Auth (Public part: login, verify-otp, password-reset)
	auth.RegisterAuthRoutes(apiV1.Group("/auth"), db)

	// 2. PROTECTED ROUTES (JWT + Tenant)
	// --------------------------------------------------
	protected := apiV1.Group("",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
	)

	// User (Internal: Profile and Staff management)
	user.RegisterUserRoutes(apiV1, protected, db)

	business.RegisterBusinessRoutes(protected, db)
	advert.RegisterAdvertRoutes(protected, db)
	business.RegisterAdminRoutes(protected, db)
	sale.RegisterManagementRoutes(protected, db)
	outlet.RegisterRoutes(protected, db)
	terminal.RegisterRoutes(protected, db)

	// 3. BUSINESS SCOPED ROUTES (JWT + Tenant + CurrentBusiness)
	// --------------------------------------------------
	businessScoped := protected.Group("",
		middleware.CurrentBusinessMiddleware(),
		middleware.SubscriptionMiddleware(db),
	)

	business.RegisterDashboardBusinessRoutes(businessScoped, db)

	category.RegisterCategoryRoutes(businessScoped, db)
	product.RegisterProductRoutes(businessScoped, db)
	inventory.RegisterInventoryRoutes(businessScoped, db)
	sale.RegisterSaleRoutes(businessScoped, db)
	expense.RegisterRoutes(businessScoped, db)
	seed.RegisterRoutes(businessScoped, db)

	// Subscriptions & Shift/Table
	subscription.RegisterRoutes(protected.Group("/subscription", middleware.CurrentBusinessMiddleware()), db)
	subscription.RegisterReferralRoutes(protected.Group("/referrals"), db)
	subscription.RegisterAdminRoutes(protected, db)

	shift.RegisterShiftRoutes(businessScoped, db)
	table.RegisterTableRoutes(businessScoped, db)
	printing.RegisterRoutes(businessScoped, db)
	recipe.RegisterRecipeRoutes(businessScoped, db)
	tutorial.RegisterRoutes(businessScoped, db)
	notification.RegisterNotificationRoutes(businessScoped, db)

	// === Start server ===
	port := config.AppPort()
	baseURL := getBaseURL()
	log.Printf("Environment: %s", getEnvironment())
	log.Printf("Server running on: %s", baseURL)
	log.Printf("Swagger UI: %s/swagger/index.html", baseURL)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}

// ---------------- Helpers ----------------

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func isDevelopment() bool {
	return strings.ToLower(getEnvironment()) == "development"
}

func getBaseURL() string {
	if isDevelopment() {
		port := config.AppPort()
		if port == "80" || port == "443" {
			return "http://localhost"
		}
		return "http://localhost:" + port
	}

	if host := os.Getenv("RENDER_EXTERNAL_HOSTNAME"); host != "" {
		return "https://" + host
	}
	if host := os.Getenv("HOSTNAME"); host != "" {
		return "https://" + host
	}
	if host := os.Getenv("ORIGIN"); host != "" {
		return host
	}
	if host := os.Getenv("API_HOST"); host != "" {
		return host
	}

	return "https://your-production-domain.com"
}

func setDynamicSwaggerConfig() {
	baseURL := getBaseURL()
	host := strings.TrimPrefix(baseURL, "http://")
	host = strings.TrimPrefix(host, "https://")

	scheme := "http"
	if strings.HasPrefix(baseURL, "https://") {
		scheme = "https"
	}

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = []string{scheme}
	docs.SwaggerInfo.BasePath = "/api/v1"

	log.Printf("Swagger configured for host: %s, scheme: %s", host, scheme)
}

func getAllowedOrigins() string {
	base := getBaseURL()
	defaults := []string{
		"https://betadaypos.vercel.app",
		"https://betadaypos-website.vercel.app",
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173",
		"http://localhost:5050",
		base,
	}

	if isDevelopment() {
		httpsVersion := strings.Replace(base, "http://", "https://", 1)
		defaults = append(defaults, httpsVersion)
	}

	// Merge environment origins if they exist
	if env := os.Getenv("ALLOWED_ORIGINS"); env != "" {
		envOrigins := strings.Split(env, ",")
		for _, o := range envOrigins {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				defaults = append(defaults, trimmed)
			}
		}
	}

	seen := make(map[string]bool)
	var result []string
	for _, origin := range defaults {
		if origin != "" && !seen[origin] {
			seen[origin] = true
			result = append(result, origin)
		}
	}
	return strings.Join(result, ",")
}
