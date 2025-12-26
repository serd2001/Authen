package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)
// 2.setup request id
func SetRequestIdMiddleware(app *fiber.App) {
	app.Use(requestid.New())
}
