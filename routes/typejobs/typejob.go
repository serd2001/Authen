package typejob

import (
	"Auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func TypeJobRoutes(router fiber.Router) {
	// TypeJobes routes
	router.Post("/register", controllers.Register)
	router.Group("/typejobs")
	router.Post("/createTypejob", controllers.CreateTypeJob)
	router.Get("/getall", controllers.GetallTypeJob)
	router.Get("/getbyid/:id", controllers.GetTypeJobByID)
	router.Delete("/delete/:id", controllers.DeleteTypeJob)
	router.Get("/search/:name", controllers.Search)
}
