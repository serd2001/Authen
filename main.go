package main

import (
	"errors"
	"log"
	"os"
	"time"

	dotenv "Auth/config"
	"Auth/database"
	"Auth/firebase"
	"Auth/middleware"
	"Auth/routes"
	"Auth/validators"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func init() {
	mode := os.Getenv("GO_ENV")
	if mode == "" {
		dotenv.SetDotenv()
	}
	log.Println("✅ .env file loaded successfully")
	//logging MODE of app
	log.Println("------ Running in '" + mode + "' mode... ------")
}
func main() {
	var apiName = os.Getenv("API_NAME")
	var apiVersion = os.Getenv("API_VERSION")
	var mode = os.Getenv("GO_ENV")
	var buildAt = os.Getenv("BUILD_DATE")
	var startRunAt = time.Now().Format("2006-01-02 15:04:05")

	database.Connect()
	// create default roles
	database.SeedRoles()
	myConfig := fiber.Config{
		AppName: apiName,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			//response error
			err = ctx.Status(code).JSON(fiber.Map{
				"timestamp": time.Now().Format("2006-01-02-15-04-05"),
				"status":    0,
				"items":     nil,
				"error":     err.Error(),
			})
			return err
		},
	}

	app := fiber.New(myConfig)
	// cors
	middleware.SetupCores(app)
	// prevent panic
	app.Use(recover.New())
	//validators
	validators.Init()
	//apply all
	app.Use(middleware.RateLimiter())
	api := app.Group("/api/" + apiVersion)
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"API_NAME":     apiName,
			"API_VERSION":  apiVersion,
			"MODE":         mode,
			"BUILD_AT":     buildAt,
			"START_RUN_AT": startRunAt,
		})
	})

	//check health status
	api.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "OK",
		})
	})
	// logger
	middleware.Setuplogger(app)
	//ramit
	app.Use(middleware.RateLimiter())
	//call firebase init
	   firebase.InitFirebase()
        
    // Check Firebase connection


    // Continue with app startup, e.g., start Fiber, routes...
	//routes.Setup(app)
	routes.SetupRoutes(app)

	//auth:api.group("/auth")
	//routes.UserDetails(app)
	//compress
	// middleware.SetCompressMiddleware(app)
	// // Create channel to signify a signal being sent
	// c := make(chan os.Signal, 1)

	// // When an interrupt or termination signal is sent, notify the channel
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	app.Listen(":3000")

	//_ = <-c // This blocks the main thread until an interrupt is received
	log.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	log.Println("Running cleanup tasks...")
	// Your cleanup tasks go here ...

	log.Println("Fiber was successful shutdown.")
}
