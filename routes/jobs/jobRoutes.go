package jobs

import (
	jobs "Auth/controllers/jobsController"

	"github.com/gofiber/fiber/v2"
)

func JobRoutes(router fiber.Router) {
	// TypeJobes routes
	router.Post("/createjob", jobs.CreateJob)
	// router.Get("/getcom", company.GetAllCompany)
	// router.Get("/getbyid/:id", company.GetCompanyByID)
	router.Put("/update/:id", jobs.UpdateJob)
	router.Delete("/delete/:id", jobs.DeleteJob)
	router.Get("/getall", jobs.GetAllJobs)
	router.Get("/getbyid/:id", jobs.GetJobByID)
	router.Get("/get", jobs.GetJobByID)
}
