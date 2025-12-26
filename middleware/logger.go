package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// 3. logger to see the request deatails
func Setuplogger(app *fiber.App){
app.Use(logger.New(logger.Config{
	Format: `{"time":"${time}","status":${status},"method":"${method}","latency":"${latency}","ip":"${ip}","path":"${path}","request_id":"${locals:requestid}","user":"${locals:user}","error":"${error}"}` + "\n",
	Output: os.Stdout,
	TimeFormat: "2006-01-02T15:04:05.000Z07:00", // ISO8601 UTC with milliseconds and timezone
	TimeZone: "UTC", // Use UTC to standardize logs across servers/timezones
	// No request/response bodies to avoid sensitive data leak and performance hit
}))
}

