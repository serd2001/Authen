package routes

import (
	auth "Auth/routes/auths"
	"Auth/routes/companies"
	jobs "Auth/routes/jobs"

	typejob "Auth/routes/typejobs"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Auth routes
	authGroup := api.Group("/auth")
	auth.AuthRoute(authGroup)

	// TypeJob routes
	typeJobGroup := api.Group("/typejob")
	typejob.TypeJobRoutes(typeJobGroup)

	// company route
	company := api.Group("/company")
	companies.CompanyRoutes(company)
	// jobs routes
	job := api.Group("/job")
	jobs.JobRoutes(job)
}
