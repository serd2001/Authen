package middleware

import (
	"Auth/firebase"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FirebaseAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		token := strings.Replace(authHeader, "Bearer ", "", 1)

		client := firebase.GetAuthClient()

		decoded, err := client.VerifyIDToken(context.Background(), token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid Firebase token",
			})
		}

		// Save user info in context
		c.Locals("uid", decoded.UID)
		c.Locals("email", decoded.Claims["email"])

		return c.Next()
	}
}
