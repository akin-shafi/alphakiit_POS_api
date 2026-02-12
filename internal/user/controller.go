package user

import (
	"fmt"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/types"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// --------------------- Create User ---------------------

// @Summary Create user
// @Description Create a user under a tenant
// @Tags Users
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body User true "User payload"
// @Success 201 {object} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func CreateUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID, _ := c.Locals("tenant_id").(string)
		if tenantID == "" {
			tenantID = c.Get("X-Tenant-ID")
		}
		if tenantID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing X-Tenant-ID"})
		}

		// Use a DTO to clearly capture all fields including password
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			Role      string `json:"role"`
			Active    bool   `json:"active"`
			OutletID  *uint  `json:"outlet_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		user := User{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  req.Password,
			Role:      req.Role,
			Active:    req.Active,
			OutletID:  req.OutletID,
			TenantID:  tenantID,
		}
		rawPassword := req.Password

		// --- Check Subscription User Limit ---
		var biz business.Business
		// Find business for this tenant (assuming 1:1 or primary business)
		if err := service.db.Where("tenant_id = ?", tenantID).First(&biz).Error; err == nil {
			sub, _ := subscription.GetSubscriptionStatus(service.db, biz.ID)

			// Default limit (e.g. Trial or Fallback)
			limit := 2

			if sub != nil {
				// Find Plan
				for _, p := range subscription.AvailablePlans {
					if p.Type == sub.PlanType {
						limit = p.UserLimit
						break
					}
				}
			}

			// Count current active users for this tenant
			var currentCount int64
			service.db.Model(&User{}).Where("tenant_id = ? AND active = ?", tenantID, true).Count(&currentCount)

			if int(currentCount) >= limit && user.Active {
				return c.Status(403).JSON(fiber.Map{
					"error":   "User Limit Reached",
					"message": fmt.Sprintf("Your current plan allows max %d users. Please upgrade to add more staff.", limit),
				})
			}
		}
		// -------------------------------------

		if err := service.Create(&user); err != nil {
			if err.Error() == "email already in use" {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Generate OTP for verification
		code, _ := otp.GenerateOTP()
		otpEntry := otp.OTP{
			Email:     strings.ToLower(user.Email),
			Code:      code,
			Type:      otp.TypeVerification,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour), // Give them a day for first set up
			Used:      false,
		}
		service.db.Create(&otpEntry)

		// Send invitation email
		go func(targetEmail, firstName, pswd, otpCode string) {
			sender := email.NewSender(email.LoadConfig())
			subject := "Invitation to join AB-POS"
			body := fmt.Sprintf(`
				<h2>Welcome to AB-POS</h2>
				<p>Hello %s,</p>
				<p>You have been added as a staff member. Here are your login credentials:</p>
				<p><b>Email:</b> %s</p>
				<p><b>Password:</b> %s</p>
				<p><b>Verification OTP:</b> %s</p>
				<p>Please log in and enter this OTP to verify your account.</p>
			`, firstName, targetEmail, pswd, otpCode)

			_ = sender.SendCustomEmail(targetEmail, subject, body)
		}(user.Email, user.FirstName, rawPassword, code)

		return c.Status(201).JSON(user)
	}
}

// --------------------- List Users ---------------------

// @Summary List users
// @Description List users for a tenant
// @Tags Users
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {array} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [get]
func ListUsersHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID, _ := c.Locals("tenant_id").(string)
		if tenantID == "" {
			tenantID = c.Get("X-Tenant-ID")
		}
		if tenantID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing X-Tenant-ID"})
		}

		users, err := service.ListByTenant(tenantID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(users)
	}
}

// --------------------- Get User Profile ---------------------

// @Summary Get user profile
// @Description Retrieve profile of a user
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func GetUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		user, err := service.GetByID(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		return c.JSON(user)
	}
}

// --------------------- Update User ---------------------

// @Summary Update user
// @Description Update user details
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body User true "User payload"
// @Success 200 {object} User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func UpdateUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			Role      string `json:"role"`
			Active    bool   `json:"active"`
			OutletID  *uint  `json:"outlet_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		userData := User{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  req.Password,
			Role:      req.Role,
			Active:    req.Active,
			OutletID:  req.OutletID,
		}

		updatedUser, err := service.Update(id, &userData)
		if err != nil {
			if err.Error() == "email already in use" {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(updatedUser)
	}
}

// --------------------- Delete User ---------------------

// @Summary Delete user
// @Description Delete a user by ID
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func DeleteUserHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		idUint64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}
		id := uint(idUint64)

		if err := service.Delete(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendStatus(204)
	}
}

// --------------------- Reset Password ---------------------

// @Summary Reset user password
// @Description Reset a user's password
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body map[string]string true "Password payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/reset-password [post]
func ResetPasswordHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
		}

		var payload struct {
			Password string `json:"password"`
		}

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		if err := service.ResetPassword(uint(id), payload.Password); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "Password reset successfully"})
	}
}

// ProfileHandler retrieves the logged-in user's profile
// @Summary Get profile
// @Description Get profile of logged-in user
// @Tags Users
// @Success 200 {object} User
// @Security BearerAuth
// @Router /profile [get]
func ProfileHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*types.UserClaims) // JWT claims from middleware
		u, err := service.GetByID(userCtx.UserID)       // UserID is uint
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.JSON(u)
	}
}

// UpdateProfileHandler allows a user to update their own profile
func UpdateProfileHandler(service *UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*types.UserClaims)

		var req struct {
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
			BankName      string `json:"bank_name"`
			AccountNumber string `json:"account_number"`
			AccountName   string `json:"account_name"`
			Phone         string `json:"phone"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		userData := User{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			BankName:      req.BankName,
			AccountNumber: req.AccountNumber,
			AccountName:   req.AccountName,
			Phone:         req.Phone,
		}

		updatedUser, err := service.Update(userCtx.UserID, &userData)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(updatedUser)
	}
}
