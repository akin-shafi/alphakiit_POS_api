package business

import (
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/user"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ChartDataPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
	Expense float64 `json:"expense"`
}

type BusinessDetailsResponse struct {
	Business Business    `json:"business"`
	Owner    *user.User  `json:"owner"`
	Staff    []user.User `json:"staff"`
	Stats    struct {
		RevenueToday float64 `json:"revenue_today"`
		RevenueWeek  float64 `json:"revenue_week"`
		RevenueMonth float64 `json:"revenue_month"`
		TotalRevenue float64 `json:"total_revenue"`
		TotalProfit  float64 `json:"total_profit"`
		TotalExpense float64 `json:"total_expense"`
	} `json:"stats"`
	ChartData []ChartDataPoint `json:"chart_data"`
}

type AdminBusinessController struct {
	db *gorm.DB
}

func NewAdminBusinessController(db *gorm.DB) *AdminBusinessController {
	return &AdminBusinessController{db: db}
}

// GetAllBusinesses returns all businesses across all tenants
// @Summary      Get all businesses (Admin)
// @Description  Retrieve all businesses
// @Tags         Admin
// @Produce      json
// @Success      200  {array}   Business
// @Security     BearerAuth
// @Router       /admin/businesses [get]
func (ac *AdminBusinessController) GetAllBusinesses(c *fiber.Ctx) error {
	var businesses []Business

	// Preload tenant info if available (assuming Tenant model is linked, but Business struct has TenantID string)
	// We might want to join with User (owner) table to get owner details if needed.
	// For now, just list businesses.

	if err := ac.db.Order("created_at DESC").Find(&businesses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(businesses)
}

// CreateBusiness creates a new business (Admin mode - requires explicit tenant_id?)
// Or maybe this creates a business and a tenant/user?
// For now, let's assume it creates a business for an existing tenant or a new one.
// Simplest: Create business for a specific TenantID provided in body.
func (ac *AdminBusinessController) CreateBusiness(c *fiber.Ctx) error {
	var req struct {
		TenantID string              `json:"tenant_id" validate:"required"`
		Name     string              `json:"name" validate:"required"`
		Type     common.BusinessType `json:"type" validate:"required"`
		Address  string              `json:"address"`
		City     string              `json:"city"`
		Currency string              `json:"currency"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	biz := Business{
		TenantID: req.TenantID,
		Name:     req.Name,
		Type:     common.BusinessType(strings.ToUpper(string(req.Type))),
		Address:  req.Address,
		City:     req.City,
		Currency: common.Currency(strings.ToUpper(req.Currency)), // Also normalize currency
	}

	if err := ac.db.Create(&biz).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Optional: Seed (Moved to separate endpoint to avoid import cycle)
	// seed.SeedSampleData(ac.db, biz.ID, biz.Type)

	return c.Status(201).JSON(biz)
}

// UpdateBusiness updates any business found by ID
func (ac *AdminBusinessController) UpdateBusiness(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	var biz Business
	if err := ac.db.First(&biz, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Business not found"})
	}

	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Prevent unauthorized field updates if necessary, but this is Admin.
	// Filter allowed fields for safety
	updates := make(map[string]interface{})
	allowed := []string{"name", "type", "address", "city", "currency", "subscription_status", "is_seeded"}

	for _, field := range allowed {
		if val, ok := req[field]; ok {
			updates[field] = val
		}
	}

	if err := ac.db.Model(&biz).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(biz)
}

// DeactivateBusiness soft deletes or sets a flag
// We can use the existing 'DeletedAt' for soft delete, or add an query param to toggle "active" status if we had one.
// The user asked for "Deactivate business". Soft delete serves this purpose usually.
// Or we can assume SubscriptionStatus might handle suspension.
// Let's implement soft delete for now.
func (ac *AdminBusinessController) DeleteBusiness(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := ac.db.Delete(&Business{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// ResetBusinessData wipes all sales, shifts, and audit logs for a business and restores stock
func (ac *AdminBusinessController) ResetBusinessData(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	businessID := uint(id)

	// Verify business exists
	var biz Business
	if err := ac.db.First(&biz, businessID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Business not found"})
	}

	err := ac.db.Transaction(func(tx *gorm.DB) error {
		// 1. Restock logic for COMPLETED sales
		type ItemStock struct {
			ProductID uint
			Quantity  int
		}
		var items []ItemStock

		// Get summing quantities of completed sales items to restore stock
		if err := tx.Table("sale_items").
			Select("sale_items.product_id, SUM(sale_items.quantity) as quantity").
			Joins("JOIN sales ON sales.id = sale_items.sale_id").
			Where("sales.business_id = ? AND sales.status = ? AND sales.deleted_at IS NULL", businessID, "COMPLETED").
			Group("sale_items.product_id").
			Scan(&items).Error; err != nil {
			return err
		}

		// Restore stock levels
		for _, item := range items {
			// Update products table
			if err := tx.Table("products").
				Where("id = ? AND business_id = ?", item.ProductID, businessID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}

			// Update inventories table
			if err := tx.Table("inventories").
				Where("product_id = ? AND business_id = ?", item.ProductID, businessID).
				UpdateColumn("current_stock", gorm.Expr("current_stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		// 2. Perform deletions (Wipe transaction data)

		// Delete SaleActivityLog
		if err := tx.Exec("DELETE FROM sale_activity_logs WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete SaleItems
		if err := tx.Exec("DELETE FROM sale_items WHERE sale_id IN (SELECT id FROM sales WHERE business_id = ?)", businessID).Error; err != nil {
			return err
		}

		// Delete SaleSummaries
		if err := tx.Exec("DELETE FROM sale_summaries WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete Sales
		if err := tx.Exec("DELETE FROM sales WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete ShiftReadings
		if err := tx.Exec("DELETE FROM shift_readings WHERE shift_id IN (SELECT id FROM shifts WHERE business_id = ?)", businessID).Error; err != nil {
			return err
		}

		// Delete Shifts
		if err := tx.Exec("DELETE FROM shifts WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reset business data: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Business data reset successfully! All test sales cleared and stock restored.",
		"business": biz.Name,
	})
}

// GetBusinessDetails returns comprehensive stats for a single business
func (ac *AdminBusinessController) GetBusinessDetails(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	businessID := uint(id)

	var biz Business
	if err := ac.db.First(&biz, businessID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Business not found"})
	}

	var response BusinessDetailsResponse
	response.Business = biz

	// 1. Get Owner
	var owner user.User
	if err := ac.db.Where("tenant_id = ? AND role = ?", biz.TenantID, "OWNER").First(&owner).Error; err == nil {
		response.Owner = &owner
	}

	// 2. Get Staff
	ac.db.Where("tenant_id = ? AND role != ?", biz.TenantID, "OWNER").Find(&response.Staff)

	// 3. Get Stats (Direct SQL to avoid cycles)
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfToday.AddDate(0, 0, -int(startOfToday.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Revenue Today
	ac.db.Table("sales").
		Where("business_id = ? AND status = ? AND sale_date >= ?", businessID, "COMPLETED", startOfToday).
		Select("SUM(total)").Scan(&response.Stats.RevenueToday)

	// Revenue Week
	ac.db.Table("sales").
		Where("business_id = ? AND status = ? AND sale_date >= ?", businessID, "COMPLETED", startOfWeek).
		Select("SUM(total)").Scan(&response.Stats.RevenueWeek)

	// Revenue Month
	ac.db.Table("sales").
		Where("business_id = ? AND status = ? AND sale_date >= ?", businessID, "COMPLETED", startOfMonth).
		Select("SUM(total)").Scan(&response.Stats.RevenueMonth)

	// Totals (Historical + Live)
	var liveRevenue float64
	var liveProfit float64
	ac.db.Table("sales").
		Joins("JOIN sale_items ON sale_items.sale_id = sales.id").
		Where("sales.business_id = ? AND sales.status = ?", businessID, "COMPLETED").
		Select("SUM(sales.total) as revenue, (SUM(sale_items.profit) - SUM(sales.discount)) as profit").
		Scan(&struct {
			Revenue float64
			Profit  float64
		}{Revenue: liveRevenue, Profit: liveProfit})

	var archivedRevenue float64
	var archivedProfit float64
	ac.db.Table("sale_summaries").
		Where("business_id = ?", businessID).
		Select("SUM(total_sales), SUM(total_profit)").
		Row().Scan(&archivedRevenue, &archivedProfit)

	response.Stats.TotalRevenue = liveRevenue + archivedRevenue
	response.Stats.TotalProfit = liveProfit + archivedProfit

	// Total Expenses
	ac.db.Table("expenses").
		Where("business_id = ? AND deleted_at IS NULL", businessID).
		Select("SUM(amount)").Scan(&response.Stats.TotalExpense)

	// 4. Chart Data (Last 6 Months)
	sixMonthsAgo := startOfMonth.AddDate(0, -5, 0)

	// Group by month
	// query := `
	// 	SELECT 
	// 		TO_CHAR(d, 'YYYY-MM') as month,
	// 		COALESCE(SUM(s.total), 0) as revenue,
	// 		COALESCE((SELECT SUM(profit) FROM sale_items WHERE sale_id IN (SELECT id FROM sales s2 WHERE s2.tenant_id = ? AND TO_CHAR(s2.sale_date, 'YYYY-MM') = TO_CHAR(d, 'YYYY-MM') AND s2.status = 'COMPLETED')), 0) as profit,
	// 		COALESCE((SELECT SUM(amount) FROM expenses WHERE tenant_id = ? AND TO_CHAR(date, 'YYYY-MM') = TO_CHAR(d, 'YYYY-MM') AND deleted_at IS NULL), 0) as expense
	// 	FROM generate_series(?, ?, '1 month'::interval) d
	// 	LEFT JOIN sales s ON s.tenant_id = ? AND TO_CHAR(s.sale_date, 'YYYY-MM') = TO_CHAR(d, 'YYYY-MM') AND s.status = 'COMPLETED'
	// 	GROUP BY month
	// 	ORDER BY month ASC
	// `
	// Wait, Postgres generate_series/TO_CHAR might not be available if using SQLite for dev,
	// but the project seems to use Postgres based on previous queries.
	// Actually, I'll use a simpler approach if possible to be safe, but let's stick to Postgres styles used elsewhere.
	// Note: BusinessID is better than TenantID here for per-business details.

	rows, err := ac.db.Raw(`
		WITH months AS (
			SELECT generate_series(date_trunc('month', ?::timestamp), date_trunc('month', NOW()), '1 month'::interval) AS m
		)
		SELECT 
			TO_CHAR(m, 'Mon YYYY') as date,
			COALESCE((SELECT SUM(total) FROM sales WHERE business_id = ? AND status = 'COMPLETED' AND date_trunc('month', sale_date) = m), 0) as revenue,
			COALESCE((SELECT SUM(profit) FROM sale_items WHERE sale_id IN (SELECT id FROM sales WHERE business_id = ? AND status = 'COMPLETED' AND date_trunc('month', sale_date) = m)), 0) as profit,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE business_id = ? AND deleted_at IS NULL AND date_trunc('month', date) = m), 0) as expense
		FROM months
		ORDER BY m ASC
	`, sixMonthsAgo, businessID, businessID, businessID).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d ChartDataPoint // Use the defined struct
			rows.Scan(&d.Date, &d.Revenue, &d.Profit, &d.Expense)
			response.ChartData = append(response.ChartData, d)
		}
	}

	return c.JSON(response)
}

// RegisterAdminRoutes registers routes for admin business management
func RegisterAdminRoutes(r fiber.Router, db *gorm.DB) {
	ac := NewAdminBusinessController(db)

	r.Get("/admin/businesses", ac.GetAllBusinesses)
	r.Post("/admin/businesses", ac.CreateBusiness)
	r.Put("/admin/businesses/:id", ac.UpdateBusiness)
	r.Delete("/admin/businesses/:id", ac.DeleteBusiness)
	r.Post("/admin/businesses/:id/reset", ac.ResetBusinessData)
	r.Get("/admin/businesses/:id/details", ac.GetBusinessDetails)
}
