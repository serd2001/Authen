package auth

import (
	"Auth/controllers"
	"Auth/middleware"
	"Auth/test"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(router fiber.Router) {
	router.Post("/register", controllers.Register)
	router.Post("/test", test.TestClaims)
	router.Post("/register", controllers.Register)
	router.Post("/sociallogin", controllers.LoginSocialFirebase)
	router.Post("/firebase-login", controllers.LoginWithFirebase)
	router.Post("/set-new-passwordemail", controllers.ForgotPasswordByEmail)

	// Protected routes (require Firebase auth)
	router.Put("/update-user", middleware.FirebaseAuth(), controllers.UpdateProfile)
	router.Get("/GetProfile", middleware.FirebaseAuth(), controllers.GetProfile)
	router.Post("/logout", middleware.FirebaseAuth(), controllers.Logout)
	router.Delete("/deletecurrent", middleware.FirebaseAuth(), controllers.DeleteCurrentUser)

	// User details management routes
	router.Group("/user")
	router.Post("/:userId", controllers.UpdateUserDetails)
	router.Get("/:id", controllers.GetUserProfile)

}
