package jobs

import (
	"Auth/database"
	"Auth/models"
	presenters "Auth/presenter"
	"strconv"
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

// update
func UpdateJob(c *fiber.Ctx) error {
	// Parse job ID from route param
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid job ID",
		})
	}

	// Find existing job
	var job models.Job
	if err := database.DB.First(&job, uint(jobID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Job not found",
		})
	}

	// Parse JSON body
	type Req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SalaryStart int64  `json:"salary_start"`
		SalaryEnd   int64  `json:"salary_end"`
		Type        string `json:"type"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		JobTypeID   uint   `json:"job_type_id"`
		CompanyID   uint   `json:"company_id"`
	}

	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}

	// Update only provided fields
	if strings.TrimSpace(req.Name) != "" {
		job.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Description) != "" {
		job.Description = strings.TrimSpace(req.Description)
	}
	if req.SalaryStart > 0 {
		job.SalaryStart = req.SalaryStart
	}
	if req.SalaryEnd > 0 {
		job.SalaryEnd = req.SalaryEnd
	}
	if strings.TrimSpace(req.Type) != "" {
		job.Type = strings.TrimSpace(req.Type)
	}

	// Date validation
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid start_date format (YYYY-MM-DD)"})
		}
		job.StartDate = t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid end_date format (YYYY-MM-DD)"})
		}
		if !job.StartDate.IsZero() && t.Before(job.StartDate) {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "EndDate must be after StartDate"})
		}
		job.EndDate = t
	}

	// Update relations if provided
	if req.CompanyID > 0 {
		var company models.Company
		if err := database.DB.First(&company, req.CompanyID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"success": false, "message": "Company not found"})
		}
		job.CompanyID = req.CompanyID
	}
	if req.JobTypeID > 0 {
		var jobType models.JobType
		if err := database.DB.First(&jobType, req.JobTypeID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"success": false, "message": "Job type not found"})
		}
		job.JobTypeID = req.JobTypeID
	}

	// Save changes
	if err := database.DB.Save(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update job"})
	}

	return c.Status(200).JSON(presenters.ResponseSuccess(job))
}

// DeleteJob
func DeleteJob(c *fiber.Ctx) error {
	// Parse job ID from route param
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid job ID",
		})
	}

	// Find job
	var job models.Job
	if err := database.DB.First(&job, uint(jobID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Job not found",
		})
	}

	// Delete job
	if err := database.DB.Delete(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete job",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Job deleted successfully",
	})
}

// getall
// get all
func GetAllJobs(c *fiber.Ctx) error {
	var jobs []models.Job

	// Pagination parameters
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	limit := c.QueryInt("limit", 30)
	if limit < 1 || limit > 100 {
		limit = 30
	}

	offset := (page - 1) * limit

	// Get total count
	var totalItems int64
	if err := database.DB.Model(&models.Job{}).Count(&totalItems).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count jobs",
		})
	}

	// Get paginated jobs (with relations if needed)
	if err := database.DB.
		Preload("Company").
		Preload("JobType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch jobs",
		})
	}

	// Calculate pagination values
	currentPage := page
	currentPageTotalItem := len(jobs)
	totalPage := int((totalItems + int64(limit) - 1) / int64(limit)) // ceiling division

	return c.Status(200).JSON(presenters.ResponseSuccessListData(
		jobs,
		currentPage,
		currentPageTotalItem,
		totalPage,
	))
}

// get by id
func GetJobByID(c *fiber.Ctx) error {
	// Parse job ID from route param
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid job ID",
		})
	}

	// Find job with relations
	var job models.Job
	if err := database.DB.
		Preload("Company").
		Preload("JobType").
		First(&job, uint(jobID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Job not found",
		})
	}

	// Success response
	return c.Status(200).JSON(presenters.ResponseSuccess(job))
}

// Define a struct for the result

// Query with joins
func GetJobs(c *fiber.Ctx) error {
	var reports []models.Job

	err := database.DB.Table("jobs").
		Select(`jobs.id as job_id,
                jobs.name as job_name,
                companies.name as company_name,
                job_types.name as job_type_name,
                jobs.salary_start,
                jobs.salary_end`).
		Joins("JOIN companies ON companies.id = jobs.company_id").
		Joins("JOIN job_types ON job_types.id = jobs.job_type_id").
		Scan(&reports).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch job report",
		})
	}

	return c.Status(200).JSON(presenters.ResponseSuccess(reports))
}
