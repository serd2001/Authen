package middleware

import (
	"Auth/database"
	"Auth/firebase"
	"Auth/models"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FirebaseAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Missing Authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid Authorization format",
			})
		}

		idToken := parts[1]

		client := firebase.GetAuthClient()
		decoded, err := client.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid Firebase token",
			})
		}

		// 🔥 Find user in DB by Firebase UID
		var user models.User
		if err := database.DB.
			Where("firebase_uid = ?", decoded.UID).
			First(&user).Error; err != nil {

			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "User not registered",
			})
		}

		// ✅ SET WHAT YOUR HANDLER EXPECTS
		c.Locals("user_id", user.ID)
		c.Locals("user", user.User_Details.Name)
		c.Locals("roles", user.Roles)
		c.Locals("firebase_uid", decoded.UID)
		c.Locals("email", user.Email)
		return c.Next()
	}
}
