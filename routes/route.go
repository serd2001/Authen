package routes

import (
	"Auth/controllers"
	"Auth/test"

	"github.com/gofiber/fiber/v2"
)

// func AuthRoutes(api fiber.Router) {
// 	auth := api.Group("/auth")

// 	// Public routes
// 	auth.Post("/signup", controllers.Registertest)
// 	auth.Post("/login", controllers.Login)
// 	auth.Post("/login/firebase", controllers.LoginWithFirebase)
// 	//auth.Post("/refresh", controllers.RefreshToken)
// 	//auth.Post("/verify", controllers.VerifyToken)
// }

// // UserRoutes sets up user routes
// func UserRoutes(api fiber.Router) {
// 	users := api.Group("/users")

// 	// Protected routes (require authentication)
// 	//users.Use(middleware.AuthMiddleware)

// 	users.Get("/profile", controllers.GetUserProfile)
// 	//users.Put("/profile", controllers.UpdateUserProfile)
// 	//users.Delete("/account", controllers.DeleteAccount)
// 	//users.Post("/change-password", controllers.ChangePassword)

// 	// Admin only routes
// 	//admin := users.Group("/admin")
// 	//admin.Use(middleware.AdminMiddleware)
// 	//admin.Get("/list", controllers.ListUsers)
// }

func Setup(app *fiber.App) {

	//app.Post("/register", controllers.Register)
	// app.Post("/login", controllers.Login)
	//  app.Get("/profile", controller.Profile)
}

func AuthRoutes(app *fiber.App) {

	auth := app.Group("/auth")
	auth.Post("/test", test.TestClaims)
	auth.Post("/register", controllers.Registertest)
	//auth.Post("/login", controllers.Login)
	auth.Post("/firebase-login", controllers.LoginWithFirebase)
	//auth.Post("/login", controllers.Login)
	// auth.Get("/profile", middleware.FirebaseAuth(), controllers.GetProfile)
	auth.Get("/profile", controllers.GetProfile)

	//auth.Post("/create-user", controllers.CreateUser)
	userDetails := auth.Group("/user")
	userDetails.Post("/:userId", controllers.UpdateUserDetails)
	userDetails.Get("/:id", controllers.GetUserProfile)
}
