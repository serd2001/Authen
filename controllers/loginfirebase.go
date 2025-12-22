package controllers

import (
	"Auth/database"
	"Auth/firebase"
	"Auth/models"
	"Auth/utils"
	"context"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func LoginWithFirebase(c *fiber.Ctx) error {
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
	jwtToken, expiresIn, err := utils.GenerateJWTWithExpiry(user.ID, user.Email, user.FirebaseUID)
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
			"token":      jwtToken,  // JWT token for API authentication
			"expires_in": expiresIn, // Token expiration time in seconds (86400 = 24 hours)
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

// getProvider determines the authentication provider from Firebase user info
// Returns provider name like "google", "facebook", "apple", or "password"
func getProvider(user *auth.UserRecord) string {
	// Check if user has any provider information
	if len(user.ProviderUserInfo) > 0 {
		// Get the first provider ID (primary authentication method)
		providerID := user.ProviderUserInfo[0].ProviderID

		// Map Firebase provider IDs to friendly names
		switch providerID {
		case "google.com":
			return "google"
		case "facebook.com":
			return "facebook"
		case "apple.com":
			return "apple"
		default:
			return providerID // Return raw provider ID if not matched
		}
	}
	// If no provider info exists, assume email/password authentication
	return "password"
}
func generateUniqueUsername(db *gorm.DB, user *auth.UserRecord, uid string) string {
	baseUsername := ""

	// Priority 1: DisplayName
	if user.DisplayName != "" {
		cleaned := strings.ReplaceAll(user.DisplayName, " ", "_")
		baseUsername = strings.ToLower(cleaned)
	} else if user.Email != "" {
		// Priority 2: Email prefix
		parts := strings.Split(user.Email, "@")
		baseUsername = parts[0]
	} else {
		// Priority 3: Fallback
		baseUsername = "user"
	}

	// Check if username exists
	username := baseUsername
	var existingUser models.User
	err := db.Where("username = ?", username).First(&existingUser).Error

	// If username exists, add UID suffix
	if err == nil {
		username = baseUsername + "_" + uid[:8]
	}

	return username
}

// good working
// func LoginWithFirebase(c *fiber.Ctx) error {
// 	type FirebaseLoginReq struct {
// 		IdToken string `json:"id_token"`
// 	}

// 	var req FirebaseLoginReq
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	if req.IdToken == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Firebase token is required",
// 		})
// 	}

// 	authClient := firebase.GetAuthClient()
// 	token, err := authClient.VerifyIDToken(context.Background(), req.IdToken)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Invalid or expired Firebase token",
// 		})
// 	}

// 	firebaseUser, err := authClient.GetUser(context.Background(), token.UID)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Failed to get user from Firebase",
// 		})
// 	}

// 	var user models.User
// 	err = database.DB.
// 		Preload("Roles").
// 		Where("firebase_uid = ?", token.UID).
// 		First(&user).Error

// 	if err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": "Database error",
// 			})
// 		}

// 		// Load default role
// 		var role models.Role
// 		if err := database.DB.Where("name = ?", "user").First(&role).Error; err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": "Role not found",
// 			})
// 		}

// 		// ✅ Generate unique username
// 		username := generateUniqueUsername(firebaseUser, token.UID)

// 		// Create new user
// 		user = models.User{
// 			FirebaseUID: token.UID,
// 			Email:       firebaseUser.Email,
// 			Username:    username, // 👈 Use generated username
// 			Provider:    getProvider(firebaseUser),
// 			Roles:       []models.Role{role},
// 		}

// 		if err := database.DB.Create(&user).Error; err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": err.Error(),
// 			})
// 		}
// 	}

// 	// Update Firebase Custom Claims
// 	// ✅ Generate claims using helper function
// 	userClaims := firebase.GenerateUserClaims(user)

// 	// ✅ Set claims to Firebase
// 	err = authClient.SetCustomUserClaims(context.Background(), user.FirebaseUID, userClaims)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to sync user claims to Firebase",
// 		})
// 	}

// 	jwtToken, err := utils.GenerateJWT(user.ID, user.Email, user.FirebaseUID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to generate token",
// 		})
// 	}

// 	return c.Status(200).JSON(fiber.Map{
// 		"message": "Login successful",
// 		"token":   jwtToken,
// 		"user": fiber.Map{
// 			"id":       user.ID,
// 			"email":    user.Email,
// 			"username": user.Username,
// 			"uid":      user.FirebaseUID,
// 			"provider": user.Provider,
// 			"roles":    getRoleNames(user.Roles), // 👈 Return clean role names
// 		},
// 	})
// }

// func LoginWithFirebase(c *fiber.Ctx) error {
// 	type FirebaseLoginReq struct {
// 		IdToken string `json:"id_token"` // Firebase ID token from client
// 	}

// 	var req FirebaseLoginReq
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	// 1️⃣ Validate input
// 	if req.IdToken == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Firebase token is required",
// 		})
// 	}

// 	// 2️⃣ Verify Firebase token
// 	authClient := firebase.GetAuthClient()
// 	token, err := authClient.VerifyIDToken(context.Background(), req.IdToken)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Invalid or expired Firebase token",
// 		})
// 	}

// 	// 3️⃣ Get user info from Firebase
// 	firebaseUser, err := authClient.GetUser(context.Background(), token.UID)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Failed to get user from Firebase",
// 		})
// 	}

// 	// 4️⃣ Find user in database by Firebase UID
// 	var user models.User
// 	err = database.DB.
// 		Preload("Roles").
// 		Where("firebase_uid = ?", token.UID).
// 		First(&user).Error

// 	// 5️⃣ If user doesn't exist, create them
// 	if err != nil {
// 		// Load default role
// 		var role models.Role
// 		if err := database.DB.Where("name = ?", "user").First(&role).Error; err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": "Role not found",
// 			})
// 		}

// 		// Create new user
// 		user = models.User{
// 			FirebaseUID: token.UID,
// 			Email:       firebaseUser.Email,
// 			Username:    firebaseUser.DisplayName,
// 			Provider:    getProvider(firebaseUser),
// 			Roles:       []models.Role{role},
// 		}

// 		if err := database.DB.Create(&user).Error; err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error": err.Error(),
// 			})
// 		}
// 	}
// 	// 5.5️⃣ Update Firebase Custom Claims with User Roles
// 	// Collect role names into a map for Firebase claims
// 	userClaims := make(map[string]interface{})
// 	var roleNames []string
// 	for _, r := range user.Roles {
// 		roleNames = append(roleNames, r.Name)
// 		userClaims[r.Name] = true // Example: {"admin": true, "user": true}
// 	}

// 	// Persist these claims to Firebase so they appear in future ID tokens
// 	err = authClient.SetCustomUserClaims(context.Background(), user.FirebaseUID, userClaims)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to sync user claims to Firebase",
// 		})
// 	}
// 	// 6️⃣ Generate JWT token
// 	jwtToken, err := utils.GenerateJWT(user.ID, user.Email, user.FirebaseUID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to generate token",
// 		})
// 	}

// 	// 7️⃣ Return success response
// 	return c.Status(200).JSON(fiber.Map{
// 		"message": "Login successful",
// 		"token":   jwtToken,
// 		"user": fiber.Map{
// 			"id":       user.ID,
// 			"email":    user.Email,
// 			"username": user.Username,
// 			"uid":      user.FirebaseUID,
// 			"provider": user.Provider,
// 		},
// 	})
// }
