package middleware

import (
	"Auth/models"

	"github.com/gofiber/fiber/v2"
)

func RequireRole(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles := c.Locals("roles").([]models.Role)
		for _, role := range roles {
			for _, r := range requiredRoles {
				if role.Name == r {
					return c.Next()
				}
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Forbidden: insufficient permissions",
		})
	}
}
