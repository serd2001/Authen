package controllers

import (
	"Auth/database"
	"Auth/firebase"
	"Auth/models"
	"Auth/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// register a new user
func Register(c *fiber.Ctx) error {
	type Req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min>=6,max=100"`
		Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
		Name     string `json:"name" validate:"required,min=1,max=100"`
		Lastname string `json:"lastname" validate:"required,min=1,max=100"`
		Gender   string `json:"gender" validate:"required,oneof=male female other prefer_not_to_say"`
		Age      int    `json:"age" validate:"required,min=1,max=150"`
		Dob      string `json:"dob" validate:"required,datetime=2006-01-02"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}
	// ✅ Prevent duplicate email / username
	var existing models.User
	if err := database.DB.
		Where("email = ? OR username = ?", req.Email, req.Username).
		First(&existing).Error; err == nil {

		return c.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "User already exists",
		})
	}

	// ✅ Create Firebase user
	authClient := firebase.GetAuthClient()
	ctx := context.Background()

	firebaseUser, err := authClient.CreateUser(ctx,
		(&auth.UserToCreate{}).
			Email(req.Email).
			Password(req.Password),
		//	DisplayName(req.Username),
	)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ✅ Load default role
	var role models.Role
	if err := database.DB.Where("name = ?", "user").First(&role).Error; err != nil {
		authClient.DeleteUser(ctx, firebaseUser.UID)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Role not found",
		})
	}

	// ✅ Start transaction
	tx := database.DB.Begin()

	// ✅ Create user (NO password stored)
	user := models.User{
		FirebaseUID: firebaseUser.UID,
		Email:       req.Email,
		Username:    req.Username,
		Provider:    "password",
		Roles:       []models.Role{role},
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		authClient.DeleteUser(ctx, firebaseUser.UID)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ✅ Create user details
	userDetails := models.User_Details{
		Name:     req.Name,
		Lastname: req.Lastname,
		Gender:   req.Gender,
		Age:      req.Age,
		Dob:      req.Dob,
		UserID:   user.ID,
	}

	if err := tx.Create(&userDetails).Error; err != nil {
		tx.Rollback()
		authClient.DeleteUser(ctx, firebaseUser.UID)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user details",
		})
	}

	tx.Commit()

	// ✅ Reload roles
	database.DB.Preload("Roles").First(&user, user.ID)

	// ✅ Set Firebase custom claims
	claims := map[string]interface{}{
		"user_id": user.ID,
		"role":    "user",
	}

	if err := authClient.SetCustomUserClaims(ctx, user.FirebaseUID, claims); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to set custom claims",
		})
	}

	// ✅ Generate JWT
	jwtToken, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.FirebaseUID,
		"user")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
		})
	}

	// ✅ Response
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Registration successful",
		"data": fiber.Map{
			"token": jwtToken,
			"user": fiber.Map{
				"id":       user.ID,
				"uid":      user.FirebaseUID,
				"email":    user.Email,
				"username": user.Username,
				"name":     userDetails.Name,
				"lastname": userDetails.Lastname,
				"provider": user.Provider,
				"roles":    []string{"user"},
			},
		},
	})
}

// GetProfile retrieves the current authenticated user's profile
func GetProfile(c *fiber.Ctx) error {
	// Get user ID from JWT token (set by auth middleware)
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user context",
		})
	}

	// Get user from database
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// Get user details
	var userDetails models.User_Details
	if err := database.DB.Where("user_id = ?", userID).First(&userDetails).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "User details not found",
		})
	}

	// Get role name
	roleName := "user"
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	// Return profile data
	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":       user.ID,
			"uid":      user.FirebaseUID,
			"email":    user.Email,
			"username": user.Username,
			"name":     userDetails.Name,
			"lastname": userDetails.Lastname,
			"gender":   userDetails.Gender,
			"age":      userDetails.Age,
			"dob":      userDetails.Dob,
			"provider": user.Provider,
			"roles":    []string{roleName},
		},
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	// =========================
	// 1. Request DTO
	// =========================
	type UpdateRequest struct {
		Username *string `json:"username" validate:"omitempty,min=3,max=30"`
		Name     *string `json:"name" validate:"omitempty,min=1,max=100"`
		Lastname *string `json:"lastname" validate:"omitempty,min=1,max=100"`
		Gender   *string `json:"gender" validate:"omitempty,oneof=male female other prefer_not_to_say"`
		Age      *int    `json:"age" validate:"omitempty,min=1,max=150"`
		Dob      *string `json:"dob" validate:"omitempty"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	// =========================
	// 2. Validate input
	// =========================
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// =========================
	// 3. Get user context
	// =========================

	userID, ok := c.Locals("user_id").(uint)
	fmt.Println("user_id:", c.Locals("user_id"))
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})

	}

	// =========================
	// 4. Load user + roles
	// =========================
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	var userDetails models.User_Details
	err := database.DB.Where("user_id = ?", userID).First(&userDetails).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		userDetails = models.User_Details{
			UserID: userID,
		}

		if err := database.DB.Create(&userDetails).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create user details",
			})
		}
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
		})
	}

	// =========================
	// 5. Check username uniqueness
	// =========================
	if req.Username != nil && *req.Username != user.Username {
		var count int64
		database.DB.Model(&models.User{}).
			Where("username = ? AND id != ?", *req.Username, userID).
			Count(&count)

		if count > 0 {
			return c.Status(409).JSON(fiber.Map{
				"success": false,
				"message": "Username already taken",
			})
		}
	}

	// =========================
	// 6. Transaction start
	// =========================
	tx := database.DB.Begin()
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to start transaction",
		})
	}

	// =========================
	// 7. Update user table
	// =========================
	if req.Username != nil {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"username": *req.Username,
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update username",
			})
		}
		user.Username = *req.Username
	}

	// =========================
	// 8. Update details table
	// =========================
	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
		userDetails.Name = *req.Name
	}
	if req.Lastname != nil {
		updates["lastname"] = *req.Lastname
		userDetails.Lastname = *req.Lastname
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
		userDetails.Gender = *req.Gender
	}
	if req.Age != nil {
		updates["age"] = *req.Age
		userDetails.Age = *req.Age
	}
	if req.Dob != nil {
		updates["dob"] = *req.Dob
		userDetails.Dob = *req.Dob
	}

	if len(updates) > 0 {
		if err := tx.Model(&userDetails).Updates(updates).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update user details",
			})
		}
	}

	// =========================
	// 9. Commit
	// =========================
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Commit failed",
		})
	}

	// =========================
	// 10. Collect roles
	// =========================
	roles := []string{}
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}
	if len(roles) == 0 {
		roles = append(roles, "user")
	}

	// =========================
	// 11. Generate new JWT
	// =========================
	//roleName := roles[0]
	token, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.FirebaseUID,
		userDetails.Name,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
		})
	}

	// =========================
	// 12. Response
	// =========================
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data": fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":       user.ID,
				"uid":      user.FirebaseUID,
				"email":    user.Email,
				"username": user.Username,
				"name":     userDetails.Name,
				"lastname": userDetails.Lastname,
				"gender":   userDetails.Gender,
				"age":      userDetails.Age,
				"dob":      userDetails.Dob,
				"provider": user.Provider,
				"roles":    roles,
			},
		},
	})
}

// UpdateEmail - Separate function to update email only
func UpdateEmaile(c *fiber.Ctx) error {

	type EmailRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	var req EmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	userID := c.Locals("user_id").(uint)
	firebaseUID := c.Locals("firebase_uid").(string)

	// Check if email is already taken
	var existing models.User
	if err := database.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "Email already taken",
		})
	}

	// Update Firebase
	authClient := firebase.GetAuthClient()
	ctx := context.Background()

	if _, err := authClient.UpdateUser(ctx, firebaseUID, (&auth.UserToUpdate{}).Email(req.Email)); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update Firebase: " + err.Error(),
		})
	}

	// Update database
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("email", req.Email).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update database",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Email updated successfully",
	})
}

// UpdatePassword - Separate function to update password only
func ForgotPasswordByEmail(c *fiber.Ctx) error {
	// 1. Request structure
	type Request struct {
		Email string `json:"email"`
	}

	var req Request

	// 2. Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// 3. Validate email
	if req.Email == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Email is required",
		})
	}
	// 4. Get Firebase API key
	apiKey := os.Getenv("FIREBASE_API_KEY_ID")
	if apiKey == "" {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Server configuration error",
		})
	}

	// 5. Prepare Firebase payload
	payload := map[string]string{
		"requestType": "PASSWORD_RESET",
		"email":       req.Email,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to prepare request",
		})
	}

	// 6. Firebase endpoint
	url := fmt.Sprintf(
		"https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key=%s",
		apiKey,
	)

	// 7. Send request to Firebase
	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to connect to authentication service",
		})
	}
	defer resp.Body.Close()

	// 8. Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to read response",
		})
	}

	// 9. Handle Firebase errors

	if resp.StatusCode != http.StatusOK {
		var firebaseErr struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		_ = json.Unmarshal(body, &firebaseErr)

		msg := firebaseErr.Error.Message
		if msg == "" {
			msg = "Failed to send password reset email"
		}

		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": msg,
		})
	}
	// 10. Success response
	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Password reset email sent successfully",
	})
}

// DeleteCurrentUser - delete user from both database and Firebase

func DeleteCurrentUser(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - user ID not found",
		})
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	// Get database connection
	db := database.DB

	// Get Firebase Auth Client
	firebaseAuth := firebase.GetAuthClient()

	// Find the user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
		})
	}

	ctx := context.Background()

	// 1. Delete Firebase Authentication
	if user.FirebaseUID != "" {
		if err := firebaseAuth.DeleteUser(ctx, user.FirebaseUID); err != nil {
			// Log error but continue (user might already be deleted in Firebase)
			// You can add logging here if needed
		}
	}

	// 2. Delete from your database (soft delete)
	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete user from database",
		})
	}

	// Clear authentication cookie (if you use cookies)
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Account deleted successfully from all systems",
	})
}

func Logout(c *fiber.Ctx) error {
	// Clear the authentication cookie by setting it to expire immediately
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to past time
		HTTPOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: "Lax",
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}
