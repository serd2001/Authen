package controllers

import (
	"Auth/database"
	"Auth/firebase"
	"Auth/models"
	"Auth/utils"
	"context"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func LoginSocialFirebase(c *fiber.Ctx) error {
	// Define the expected request structure
	type FirebaseLoginReq struct {
		IdToken string `json:"id_token"` // Firebase ID token from client
	}

	// Parse the incoming JSON request body
	var req FirebaseLoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	// Validate that the ID token is provided
	if req.IdToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Firebase token is required",
		})
	}

	// Get Firebase Auth client instance
	authClient := firebase.GetAuthClient()

	// Verify the Firebase ID token to ensure it's valid and not expired
	token, err := authClient.VerifyIDToken(context.Background(), req.IdToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Invalid or expired Firebase token",
		})

	}

	// Retrieve the complete Firebase user profile using the UID from the verified token
	firebaseUser, err := authClient.GetUser(context.Background(), token.UID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user from Firebase",
		})
	}

	// Try to find an existing user in the database by Firebase UID
	// Preload the user's roles to avoid N+1 query issues
	var user models.User
	err = database.DB.
		Preload("Roles").
		Where("firebase_uid = ?", token.UID).
		First(&user).Error

	// Handle user not found scenario
	if err != nil {
		// If error is something other than "not found", it's a database issue
		if err != gorm.ErrRecordNotFound {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Database error",
			})
		}

		// Load the default "user" role from the database
		var role models.Role
		if err := database.DB.Where("name = ?", "user").First(&role).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}

		// Generate a unique username based on Firebase user info
		// This prevents duplicate username errors
		username := generateUniqueUsername(database.DB, firebaseUser, token.UID)

		// Create a new user record with data from Firebase
		user = models.User{
			FirebaseUID: token.UID,                 // Firebase unique identifier
			Email:       firebaseUser.Email,        // User's email address
			Username:    username,                  // Generated unique username
			Provider:    getProvider(firebaseUser), // Authentication provider (google, facebook, etc.)
			Roles:       []models.Role{role},       // Assign default "user" role
		}

		// Insert the new user into the database
		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		// creat user details
		database.DB.Create(&models.User_Details{
			UserID: user.ID,
		})
	}

	// Generate custom claims for Firebase token
	// This includes user_id and roles that will be embedded in future Firebase tokens
	userClaims := firebase.GenerateUserClaims(user)

	// Update Firebase custom claims so they appear in the user's ID token
	// This allows the client to verify user permissions without additional API calls
	err = authClient.SetCustomUserClaims(context.Background(), user.FirebaseUID, userClaims)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to sync user claims to Firebase",
		})
	}

	// Generate our application's JWT token for API authentication
	// This token will be used for subsequent API requests
	jwtToken, err := utils.GenerateJWTWithExpiry(user.ID, user.Email, user.FirebaseUID, "user")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
		})
	}

	// Return successful login response with token and user information
	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": fiber.Map{
			"token": jwtToken, // JWT token for API authentication
			//"expires_in": expiresIn, // Token expiration time in seconds (86400 = 24 hours)
			"user": fiber.Map{
				"id":       user.ID,                           // Database user ID
				"uid":      user.FirebaseUID,                  // Firebase unique identifier
				"email":    user.Email,                        // User's email address
				"provider": user.Provider,                     // Authentication provider
				"roles":    firebase.GetRoleNames(user.Roles), // Array of role names (e.g., ["user", "admin"])
			},
		},
	})
}
