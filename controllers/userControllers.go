package controllers

import (
	"Auth/database"
	"Auth/models"
	presenters "Auth/presenter"
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetUserProfile(c *fiber.Ctx) error {
	// Get and validate user ID from params
	userIDParam := c.Params("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	var user models.User

	// Query with proper type conversion
	result := database.DB.Preload("User_Details").Preload("Roles").
		First(&user, uint(userID))

	// Handle different error cases
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		// Log the actual error for debugging
		log.Printf("Database error: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve user",
		})
	}

	// Create a safe response that excludes sensitive fields
	response := fiber.Map{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"details":  user.User_Details,
		"roles":    user.Roles,
		// Don't include: Password, PasswordHash, Tokens, etc.
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"user":    response,
	})
}

func UpdateUserDetails(c *fiber.Ctx) error {
	// Get user ID from params or from context (depends on your auth flow)
	userIDParam := c.Params("userId")
	var user models.User
	if err := database.DB.First(&user, userIDParam).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Parse input JSON to UserDetails struct
	var input models.User_Details
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Check if UserDetails already exist for this user
	var userDetails models.User_Details
	err := database.DB.Where("user_id = ?", user.ID).First(&userDetails).Error
	if err != nil {
		// Not found → create new
		input.UserID = user.ID
		if err := database.DB.Create(&input).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user details"})
		}
		return c.JSON(presenters.ResponseSuccess("create success"))
	} else {
		// Found → update
		userDetails.Name = input.Name
		userDetails.Lastname = input.Lastname
		userDetails.Gender = input.Gender
		userDetails.Age = input.Age
		userDetails.Dob = input.Dob

		if err := database.DB.Save(&userDetails).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update user details",
			})
		}
		return c.JSON(presenters.ResponseSuccess("update success"))
	}
	
}
