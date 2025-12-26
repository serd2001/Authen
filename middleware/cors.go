package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

//1. CORS middleware is correclty or not
func SetupCores(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		// AllowOrigins: "*",
		// AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		// AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))
}
