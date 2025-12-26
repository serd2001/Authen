package companies

import (
	company "Auth/controllers/companies"

	"github.com/gofiber/fiber/v2"
)

func CompanyRoutes(router fiber.Router) {
	// TypeJobes routes
	router.Post("/createcom", company.CreateCompany)
	router.Get("/getcom", company.GetAllCompany)
	router.Get("/getbyid/:id", company.GetCompanyByID)
	router.Put("/update/:id", company.UpdateCompany)
	router.Delete("/delete/:id", company.DeleteCompany)
}
