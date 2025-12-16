package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"pos-fiber-app/internal/user"
)

// Role represents user role in the system
type Role string

const (
	RoleOwner   Role = "OWNER"
	RoleManager Role = "MANAGER"
	RoleCashier Role = "CASHIER"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
	OutletID *uint  `json:"outlet_id,omitempty"`
	jwt.RegisteredClaims
}

// LoginRequest is the payload for login
type LoginRequest struct {
	Email    string `json:"email" example:"admin@biz.com"`
	Password string `json:"password" example:"password"`
}

// Login godoc
// @Summary Login
// @Description Authenticate a user and return access + refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/login [post]
func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
		}

		var u user.User
		if err := db.Where("email = ? AND active = true", req.Email).First(&u).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}

		if !user.CheckPassword(u.Password, req.Password) {
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}

		claims := Claims{
			UserID:   u.ID,
			TenantID: u.TenantID,
			Role:     Role(u.Role), // cast string to Role
			OutletID: u.OutletID,
		}

		access, _ := GenerateAccessToken(claims)
		refresh, _ := GenerateRefreshToken(claims)

		// Persist refresh token in DB
		db.Create(&RefreshToken{
			UserID:    u.ID,
			Token:     refresh,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})

		return c.JSON(fiber.Map{
			"access_token":  access,
			"refresh_token": refresh,
		})
	}
}

// Refresh godoc
// @Summary Refresh Access Token
// @Description Validate refresh token and issue new access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Refresh token payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/refresh [post]
func Refresh(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: validate refresh token, issue new access token
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}

// Logout godoc
// @Summary Logout
// @Description Delete refresh token and logout user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 204
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func Logout(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*Claims)
		db.Where("user_id = ?", userCtx.UserID).Delete(&RefreshToken{})
		return c.SendStatus(204)
	}
}
