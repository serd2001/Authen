package company

import (
	"Auth/database"
	"Auth/models"
	presenters "Auth/presenter"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// get all company
func GetAllCompany(c *fiber.Ctx) error {
	var companies []models.Company

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
	if err := database.DB.Model(&models.Company{}).Count(&totalItems).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜິດພາດໃນການດຶງຂໍ້ມູນ",
		})
	}

	// Get paginated data
	if err := database.DB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&companies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "ຜິດພາດໃນການດຶງຂໍ້ມູນ",
		})
	}

	// Calculate pagination values
	currentPage := page
	currentPageTotalItem := len(companies)
	totalPage := int((totalItems + int64(limit) - 1) / int64(limit)) // ceiling division

	return c.Status(200).JSON(presenters.ResponseSuccessListData(
		companies,
		currentPage,
		currentPageTotalItem,
		totalPage,
	))
}

// get by id
func GetCompanyByID(c *fiber.Ctx) error {
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
	var company models.Company

	// query from database
	if err := database.DB.
		Where("id = ? AND status = ?", id, 1).
		First(&company).Error; err != nil {

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

	return c.Status(200).JSON(presenters.ResponseSuccess(company))
}

// create typjob
func CreateCompany(c *fiber.Ctx) error {
	// Parse form fields
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	address := strings.TrimSpace(c.FormValue("address"))
	description := strings.TrimSpace(c.FormValue("description"))

	// Validate required fields
	if name == "" || email == "" || address == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Name, Email, and Address are required",
		})
	}

	// Handle logo upload
	var logoPath string
	file, err := c.FormFile("logo")
	if err == nil {
		// Save file locally (you can replace with cloud storage logic)
		savePath := fmt.Sprintf("./uploads/logos/%s", file.Filename)
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to save logo",
			})
		}
		logoPath = savePath
	}

	// Create company model
	company := models.Company{
		Name:        name,
		Email:       email,
		Address:     address,
		Description: description,
		Status:      1,
		Logo:        logoPath,
	}

	// Insert into DB

	if err := database.DB.Create(&company).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Email already exists",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save company",
		})
	}

	// Success response
	return c.Status(201).JSON(presenters.ResponseSuccess(company))

}

// update
func UpdateCompany(c *fiber.Ctx) error {
	companyID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid company ID",
		})
	}

	var company models.Company
	if err := database.DB.First(&company, companyID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Company not found",
		})
	}

	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	address := strings.TrimSpace(c.FormValue("address"))
	description := strings.TrimSpace(c.FormValue("description"))

	if name != "" {
		company.Name = name
	}
	if email != "" {
		company.Email = email
	}
	if address != "" {
		company.Address = address
	}
	if description != "" {
		company.Description = description
	}

	file, err := c.FormFile("logo")
	if file != nil {
		os.MkdirAll("./uploads/logos", 0755)
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads", "logos", filename)

		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to save logo"})
		}
		company.Logo = savePath
	}

	// Save instead of Updates
	if err := database.DB.Save(&company).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update company"})
	}

	return c.JSON(presenters.ResponseSuccess(company))
}

// delete
func DeleteCompany(c *fiber.Ctx) error {
	companyID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid company ID",
		})
	}

	var company models.Company
	if err := database.DB.First(&company, companyID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Company not found",
		})
	}

	company.Status = 0
	database.DB.Save(&company)
	if err := database.DB.Delete(&company).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete company",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Company removed successfully",
	})
}
