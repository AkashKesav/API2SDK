package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// GetUserID returns a default user ID since authentication is disabled
func GetUserID(c fiber.Ctx) (string, bool) {
	// Return a default user ID for all requests
	return "default-user", true
}

// NoAuthMiddleware is a pass-through middleware that sets a default user context
func NoAuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set a default user ID in the context
		c.Locals("user_id", "default-user")
		c.Locals("user_role", "user")
		return c.Next()
	}
}
