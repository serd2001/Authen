package test

import (
	"Auth/firebase"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func TestClaims(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Missing token"})
	}

	tokenStr := strings.Replace(authHeader, "Bearer ", "", 1)

	authClient := firebase.GetAuthClient()
	token, err := authClient.VerifyIDToken(context.Background(), tokenStr)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	return c.JSON(fiber.Map{
		"uid":    token.UID,
		"claims": token.Claims,
	})
}
