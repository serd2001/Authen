package jobs

import (
	"Auth/database"
	"Auth/models"
	presenters "Auth/presenter"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// create
func CreateJob(c *fiber.Ctx) error {
	type Req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SalaryStart int64  `json:"salary_start"`
		SalaryEnd   int64  `json:"salary_end"`
		Type        string `json:"type"`       // e.g. "full_time", "part_time"
		StartDate   string `json:"start_date"` // format "YYYY-MM-DD"
		EndDate     string `json:"end_date"`   // format "YYYY-MM-DD"
		JobTypeID   uint   `json:"job_type_id"`
		CompanyID   uint   `json:"company_id"`
	}

	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	// Trim and validate required fields
	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)

	if req.Name == "" || req.Type == "" || req.JobTypeID == 0 || req.CompanyID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Name, Type, JobTypeID, and CompanyID are required",
		})
	}

	// Salary validation
	if req.SalaryStart < 0 || req.SalaryEnd < 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Salary must be non-negative",
		})
	}
	if req.SalaryEnd > 0 && req.SalaryStart > req.SalaryEnd {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "SalaryStart cannot be greater than SalaryEnd",
		})
	}

	// Date parsing and validation
	var startDate, endDate time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Invalid start_date format (YYYY-MM-DD)",
			})
		}
		startDate = t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Invalid end_date format (YYYY-MM-DD)",
			})
		}
		endDate = t
		if !startDate.IsZero() && endDate.Before(startDate) {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "EndDate must be after StartDate",
			})
		}
	}

	// Check if company exists
	var company models.Company
	if err := database.DB.First(&company, req.CompanyID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Company not found",
		})
	}

	// Check if job type exists
	var jobType models.JobType
	if err := database.DB.First(&jobType, req.JobTypeID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Job type not found",
		})
	}

	// Create Job record
	job := models.Job{
		Name:        req.Name,
		Description: req.Description,
		SalaryStart: req.SalaryStart,
		SalaryEnd:   req.SalaryEnd,
		Type:        req.Type,
		StartDate:   startDate,
		EndDate:     endDate,
		JobTypeID:   req.JobTypeID,
		CompanyID:   req.CompanyID,
	}

	if err := database.DB.Create(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save job",
		})
	}

	// Reload job with JobType and Company using Joins
	if err := database.DB.
		Joins("JobType").
		Joins("Company").
		First(&job, job.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to load job relations",
		})
	}

	// Return full job with relationships
	return c.Status(201).JSON(presenters.ResponseSuccess(job))
}
