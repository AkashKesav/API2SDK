package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// AdminRequired is a pass-through middleware since authentication is disabled
func AdminRequired() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set admin role for all requests since authentication is disabled
		c.Locals("user_role", "admin")
		return c.Next()
	}
}
