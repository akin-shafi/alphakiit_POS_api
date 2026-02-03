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

	// NEW: Modern Swaggo for Fiber v2/v3 + OpenAPI 3

	"pos-fiber-app/docs" // Keep this for swag init to generate docs
	"pos-fiber-app/internal/auth"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/config"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/middleware"
	"pos-fiber-app/internal/onboarding"
	"pos-fiber-app/internal/outlet"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/seed"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/terminal"
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
// @description                 JWT Authorization header using the Bearer scheme. Enter your token only (without "Bearer " prefix).

func main() {
	config.LoadEnv()

	db := database.ConnectDB()

	// if err := database.RunMigrations(db); err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }

	app := fiber.New()

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
		AllowOrigins: getAllowedOrigins(),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
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

	// 1. Public Routes
	// Onboarding (Public)
	onboarding.RegisterRoutes(apiV1, db)

	// Auth (mostly Public, Logout is Protected inside RegisterAuthRoutes)
	auth.RegisterAuthRoutes(apiV1.Group("/auth"), db)

	// User (Public part: create user)
	// We pass the public group and a specially created protected user group
	protectedUserGroup := apiV1.Group("", middleware.JWTProtected(), middleware.TenantMiddleware())
	user.RegisterUserRoutes(apiV1, protectedUserGroup, db)

	// 2. Protected Routes (JWT + Tenant)
	protected := apiV1.Group("",
		middleware.JWTProtected(),
		middleware.TenantMiddleware(),
	)

	business.RegisterBusinessRoutes(protected, db)
	outlet.RegisterRoutes(protected, db)
	terminal.RegisterRoutes(protected, db)

	// 3. Business Scoped Routes (JWT + Tenant + CurrentBusiness)
	businessScoped := protected.Group("",
		middleware.CurrentBusinessMiddleware(),
		middleware.SubscriptionMiddleware(db),
	)

	category.RegisterCategoryRoutes(businessScoped, db)
	product.RegisterProductRoutes(businessScoped, db)
	inventory.RegisterInventoryRoutes(businessScoped, db)
	sale.RegisterSaleRoutes(businessScoped, db)
	subscription.RegisterRoutes(businessScoped, db)
	seed.RegisterRoutes(businessScoped, db)

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
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:5050",
		base,
	}

	if isDevelopment() {
		httpsVersion := strings.Replace(base, "http://", "https://", 1)
		defaults = append(defaults, httpsVersion)
	}

	if env := os.Getenv("ALLOWED_ORIGINS"); env != "" {
		return env
	}

	seen := make(map[string]bool)
	var result []string
	for _, origin := range defaults {
		if !seen[origin] {
			seen[origin] = true
			result = append(result, origin)
		}
	}
	return strings.Join(result, ",")
}
