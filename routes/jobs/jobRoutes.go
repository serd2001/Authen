package jobs

import (
	"Auth/controllers/jobsController"

	"github.com/gofiber/fiber/v2"
)

func JobRoutes(router fiber.Router) {
	// TypeJobes routes
	router.Post("/createjob", jobs.CreateJob)
	// router.Get("/getcom", company.GetAllCompany)
	// router.Get("/getbyid/:id", company.GetCompanyByID)
	// router.Put("/update/:id", company.UpdateCompany)
	// router.Delete("/delete/:id", company.DeleteCompany)
}
