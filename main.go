// package main

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"

// 	"Auth/database"
// 	"Auth/firebase"
// 	"Auth/routes"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	"github.com/gofiber/fiber/v2/middleware/logger"
// 	"github.com/gofiber/fiber/v2/middleware/recover"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// 1️⃣ Load environment variables
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("⚠️  No .env file found, using system environment variables")
// 	} else {
// 		log.Println("✅ .env file loaded successfully")
// 	}

// 	// 2️⃣ Connect to database
// 	log.Println("🔌 Connecting to database...")
// 	database.Connect()
// 	log.Println("✅ Database connected")

// 	// 3️⃣ Seed default roles
// 	log.Println("🌱 Seeding roles...")
// 	database.SeedRoles()
// 	log.Println("✅ Roles seeded")

// 	// 4️⃣ Initialize Firebase
// 	log.Println("🔥 Initializing Firebase...")
// 	firebase.InitFirebase()

// 	authClient := firebase.GetAuthClient()
// 	if authClient == nil {
// 		log.Fatal("❌ Failed to initialize Firebase Auth Client")
// 	}
// 	log.Println("✅ Firebase initialized successfully")

// 	// 5️⃣ Initialize Fiber app
// 	app := fiber.New(fiber.Config{
// 		AppName:      "Your App Name",
// 		ErrorHandler: customErrorHandler,
// 	})

// 	// 6️⃣ Setup middleware
// 	setupMiddleware(app)
// 	routes.AuthRoutes(app)
// 	// 7️⃣ Setup routes
// 	//setupRoutes(app)

// 	// 8️⃣ Get port from environment
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "3000"
// 	}

// 	// 9️⃣ Start server with graceful shutdown
// 	go func() {
// 		log.Printf("🚀 Server starting on http://localhost:%s", port)
// 		if err := app.Listen(":" + port); err != nil {
// 			log.Fatalf("❌ Server failed to start: %v", err)
// 		}
// 	}()

// 	// 🔟 Wait for interrupt signal for graceful shutdown
// 	gracefulShutdown(app)
// }

// // setupMiddleware configures all middleware
// func setupMiddleware(app *fiber.App) {
// 	// Recover from panics
// 	app.Use(recover.New())

// 	// Logger middleware
// 	app.Use(logger.New(logger.Config{
// 		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
// 		TimeFormat: "15:04:05",
// 		TimeZone:   "Local",
// 	}))

// 	// CORS middleware
// 	app.Use(cors.New(cors.Config{
// 		AllowOrigins: "*",
// 		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
// 		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
// 	}))
// }

// // setupRoutes registers all application routes
// // func setupRoutes(app *fiber.App) {
// // 	// Health check
// // 	app.Get("/", func(c *fiber.Ctx) error {
// // 		return c.JSON(fiber.Map{
// // 			"status":  "ok",
// // 			"message": "Server is running",
// // 			"version": "1.0.0",
// // 		})
// // 	})

// // 	// API routes
// // //	api := app.Group("/api")

// // 	// Auth routes
// // 	routes.AuthRoutes()

// // 	// User routes
// // 	//routes.UserRoutes(api)
// // }

// // setupTestRoutes creates test endpoints (development only)
// func setupTestRoutes(api fiber.Router) {
// 	test := api.Group("/test")

// 	// Generate custom Firebase token
// 	test.Get("/custom-token", GetCustomTokenHandler)

// 	// Test Firebase connection
// 	test.Get("/firebase", func(c *fiber.Ctx) error {
// 		authClient := firebase.GetAuthClient()
// 		if authClient == nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": "Firebase not initialized",
// 			})
// 		}
// 		return c.JSON(fiber.Map{
// 			"status":  "ok",
// 			"message": "Firebase is connected",
// 		})
// 	})
// }

// // GetCustomTokenHandler generates a custom Firebase token
// func GetCustomTokenHandler(c *fiber.Ctx) error {
// 	uid := c.Query("uid")
// 	if uid == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "uid parameter is required",
// 		})
// 	}

// 	authClient := firebase.GetAuthClient()
// 	if authClient == nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Firebase not initialized",
// 		})
// 	}

// 	token, err := authClient.CustomToken(context.Background(), uid)
// 	if err != nil {
// 		log.Printf("Error creating custom token: %v", err)
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to create custom token",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success":      true,
// 		"custom_token": token,
// 		"uid":          uid,
// 	})
// }

// // customErrorHandler handles application errors
// func customErrorHandler(c *fiber.Ctx, err error) error {
// 	code := fiber.StatusInternalServerError
// 	message := "Internal Server Error"

// 	if e, ok := err.(*fiber.Error); ok {
// 		code = e.Code
// 		message = e.Message
// 	}

// 	log.Printf("Error: %v", err)

// 	return c.Status(code).JSON(fiber.Map{
// 		"error":   true,
// 		"message": message,
// 	})
// }

// // gracefulShutdown handles graceful shutdown on interrupt signals
// func gracefulShutdown(app *fiber.App) {
// 	// Create channel to listen for signals
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

// 	// Block until signal received
// 	<-quit
// 	log.Println("\n🛑 Shutting down server...")

// 	// Shutdown Fiber app
// 	if err := app.Shutdown(); err != nil {
// 		log.Fatalf("❌ Server forced to shutdown: %v", err)
// 	}

// 	// Close database connection
// 	database.Close()

// 	log.Println("✅ Server exited properly")
// }

package main

import (
	"log"

	"Auth/database"
	"Auth/firebase"
	"Auth/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ .env file NOT loaded")
	}
	log.Println("✅ .env file loaded successfully")

	database.Connect()
	// create default roles
	database.SeedRoles()
	app := fiber.New()

	//call firebase init
	firebase.InitFirebase()
	authClient := firebase.GetAuthClient()
	if authClient != nil {
		log.Println("create firebase success")
	}
	// test firebase
	// Register routes
	//routes.Setup(app)
	routes.AuthRoutes(app)
	//routes.UserDetails(app)
	app.Listen(":3000")
}
