package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// GlobalLimiter returns a Fiber middleware to apply a global rate limit
func GlobalLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later.",
			})
		},
	})
}

// LoginLimiter returns a Fiber middleware to limit login attempts per email/IP
func LoginLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			type LoginPayload struct {
				Email string `json:"email"`
			}
			var payload LoginPayload
			if err := c.BodyParser(&payload); err == nil && payload.Email != "" {
				return payload.Email
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many login attempts, please try again after 1 minute.",
			})
		},
	})
}
