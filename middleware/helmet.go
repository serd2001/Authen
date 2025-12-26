package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)
//4.secrurity protection from web attacks
func SetHelmetMiddleware(app *fiber.App) {
	app.Use(helmet.New())
}
