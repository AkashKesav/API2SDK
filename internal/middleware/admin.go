package middleware

import (
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/gofiber/fiber/v3"
)

// AdminRequired protects routes that require admin privileges.
// It expects JWTMiddleware to have run first and set user context.
func AdminRequired() fiber.Handler {
	return func(c fiber.Ctx) error {
		user, exists := GetUser(c)
		if !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
				"error":   "no_user_context",
			})
		}

		if user.Role != string(models.RoleAdmin) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success":       false,
				"message":       "Admin access required",
				"error":         "insufficient_permissions",
				"required_role": string(models.RoleAdmin),
				"user_role":     user.Role,
			})
		}

		return c.Next()
	}
}
