package typejob

import (
	"Auth/controllers"
	jobtypes "Auth/controllers/jobsController"

	"github.com/gofiber/fiber/v2"
)

func TypeJobRoutes(router fiber.Router) {
	// TypeJobes routes
	router.Post("/register", controllers.Register)
	router.Group("/typejobs")
	router.Post("/createTypejob", jobtypes.CreateJob)
	router.Get("/getall", jobtypes.GetAllJobs)
	router.Get("/getbyid/:id", jobtypes.GetJobByID)
	router.Delete("/delete/:id", jobtypes.DeleteJob)
	//router.Get("/search/:name", jobtypes.)
}
