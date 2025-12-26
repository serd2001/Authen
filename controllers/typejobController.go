package controllers

import (
	"Auth/database"
	"Auth/models"
	presenters "Auth/presenter"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// get all job types
// get all job types
func GetallTypeJob(c *fiber.Ctx) error {
	var jobTypes []models.JobType

	// Get pagination parameters from query
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	limit := c.QueryInt("limit", 50)
	if limit < 1 || limit > 100 {
		limit = 30
	}

	offset := (page - 1) * limit

	// Get total count
	var totalItems int64
	if err := database.DB.Model(&models.JobType{}).Where("status = ?", 1).Count(&totalItems).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜີດພາດໃນການດຶງຂໍ້ມູນ",
		})
	}

	// Get paginated data
	if err := database.DB.
		Where("status = ?", 1).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobTypes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜີດພາດໃນການດຶງຂໍ້ມູນ",
		})
	}

	// Calculate pagination values
	currentPage := page
	currentPageTotalItem := len(jobTypes)                            // Items in current page
	totalPage := int((totalItems + int64(limit) - 1) / int64(limit)) // Ceiling division

	return c.Status(200).JSON(presenters.ResponseSuccessListData(
		jobTypes,
		currentPage,
		currentPageTotalItem,
		totalPage,
	))
}

// getById
// Get job type by ID
func GetTypeJobByID(c *fiber.Ctx) error {
	// get id
	id := c.Params("id")

	// check id int or not

	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID ຕ້ອງເປັນຕົວເລກ",
		})
	}
	var jobType models.JobType

	// query from database
	if err := database.DB.
		Where("id = ? AND status = ?", id, 1).
		First(&jobType).Error; err != nil {

		// if not see
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "ບໍ່ພົບຂໍ້ມູນປະເພດນີ້",
			})
		}

		// else err
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜີດພາດໃນການດຶງຂໍ້ມູນ",
		})
	}

	return c.Status(200).JSON(presenters.ResponseSuccess(jobType))
}

// create typjob
func CreateTypeJob(c *fiber.Ctx) error {
	type Req struct {
		Name string `json:"name" validate:"required"`
	}

	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	// Sanitize
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ກາລຸນາປ້ອນຊື່",
		})
	}

	// Rest of your code...
	jobType := models.JobType{
		Name:   req.Name,
		Status: 1,
	}

	if err := database.DB.Create(&jobType).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "ຊື່ປະເພດນີ້ ໄດ້ຖືກໃຊ້ແລ້ວ",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜີດພາດໃນການບັນທືກຂໍ້ມູນ",
		})
	}

	return c.Status(201).JSON(presenters.ResponseSuccess(jobType))
}

// CreateTypeJob creates a new job type
func UpdateTypeJob(c *fiber.Ctx) error {
	// Get and validate ID
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		return c.Status(404).JSON(presenters.ResponseError(c, 400, "ID ຕ້ອງເປັນຕົວເລກທີ່ຖືກຕ້ອງ"))
	}

	// Parse request body
	type Req struct {
		Name string `json:"name" validate:"required,min=2,max=100"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(404).JSON(presenters.ResponseError(c, 400, "ຂໍ້ມູນທີ່ສົ່ງມາບໍ່ຖືກຕ້ອງ"))
	}

	// Sanitize
	req.Name = strings.TrimSpace(req.Name)

	// Check if exists
	var jobType models.JobType
	if err := database.DB.Where("id = ? AND status = ?", idInt, models.StatusActive).First(&jobType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(presenters.ResponseError(c, 404, "ບໍ່ພົບຂໍ້ມູນປະເພດນີ້"))
		}
		return c.Status(404).JSON(presenters.ResponseError(c, 500, "ຜີດພາດໃນການດຶງຂໍ້ມູນ"))
	}

	// Skip if no change
	if jobType.Name == req.Name {
		return c.Status(200).JSON(presenters.ResponseSuccess(jobType))
	}

	// Update
	jobType.Name = req.Name
	if err := database.DB.Save(&jobType).Error; err != nil {
		// Use DB-specific error handling instead of string matching
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return presenters.ResponseError(c, 400, "ຊື່ປະເພດນີ້ ໄດ້ຖືກໃຊ້ແລ້ວ")
		}
		return presenters.ResponseError(c, 500, "ຜີດພາດໃນການບັນທືກຂໍ້ມູນ")
	}

	return c.Status(200).JSON(presenters.ResponseSuccess(jobType))
}

// DeleteTypeJob soft deletes a job type by ID
func DeleteTypeJob(c *fiber.Ctx) error {
	// 1. Validate ID
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(
			presenters.ResponseError(c, fiber.StatusBadRequest, "ID ຕ້ອງເປັນຕົວເລກທີ່ຖືກຕ້ອງ"),
		)
	}

	// 2. Find record including deleted ones
	var jobType models.JobType
	if err := database.DB.Unscoped().First(&jobType, idInt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				presenters.ResponseError(c, fiber.StatusNotFound, "ບໍ່ພົບຂໍ້ມູນປະເພດນີ້"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			presenters.ResponseError(c, fiber.StatusInternalServerError, "ຜີດພາດໃນການດຶງຂໍ້ມູນ"),
		)
	}

	// 3. Already deleted check
	if jobType.DeletedAt.Valid {
		return c.Status(fiber.StatusConflict).JSON(
			presenters.ResponseError(c, fiber.StatusConflict, "ຂໍ້ມູນ ID ນີ້ ໄດ້ຖືກລົບແລ້ວ"),
		)
	}
	// 5. Soft delete
	jobType.Status = 0
	if err := database.DB.Delete(&jobType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			presenters.ResponseError(c, fiber.StatusInternalServerError, "ຜີດພາດໃນການລົບຂໍ້ມູນ"),
		)
	}

	// 6. Reload to get DeletedAt timestamp
	jobType.Status = 0
	if err := database.DB.Unscoped().First(&jobType, idInt).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			presenters.ResponseError(c, fiber.StatusInternalServerError, "ຜີດພາດໃນການດຶງຂໍ້ມູນຫຼັງການລົບ"),
		)
	}

	// 7. Success response with deleted_at
	return c.Status(fiber.StatusOK).JSON(presenters.ResponseSuccess(fiber.Map{
		"message":    "ລົບຂໍ້ມູນສຳເລັດ",
		"id":         idInt,
		"deleted_at": jobType.DeletedAt.Time.Format(time.RFC3339),
	}))
}

// Search finds job types by name (case-insensitive, partial match)
func Search(c *fiber.Ctx) error {
	// 1. Extract and sanitize input
	name := strings.TrimSpace(c.Params("name"))
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			presenters.ResponseError(c, fiber.StatusBadRequest, "ຕ້ອງລະບຸຊື່ເພື່ອຄົ້ນຫາ"),
		)
	}

	// 2. Query database (exclude soft-deleted records)
	var jobTypes []models.JobType
	if err := database.DB.
		Where("LOWER(name) LIKE LOWER(?) AND deleted_at IS NULL", "%"+name+"%").
		Find(&jobTypes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			presenters.ResponseError(c, fiber.StatusInternalServerError, "ຜີດພາດໃນການຄົ້ນຫາ"),
		)
	}

	// 3. Handle no results
	if len(jobTypes) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			presenters.ResponseError(c, fiber.StatusNotFound, "ບໍ່ພົບຂໍ້ມູນທີ່ກົງກັບຊື່"),
		)
	}

	// 4. Return success with results
	return c.Status(fiber.StatusOK).JSON(presenters.ResponseSuccess(jobTypes))
}
