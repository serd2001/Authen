package controllers

import (
	"Auth/database"
	"Auth/firebase"
	"Auth/models"
	"Auth/utils"
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// / test register
// Registertest handles new user registration with Firebase and database
// Creates Firebase account, stores user in database, and sets custom claims for roles

func Registertest(c *fiber.Ctx) error {
	type Req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Username string `json:"username" validate:"required"`
		Name     string `json:"name" validate:"required"`
		Lastname string `json:"lastname" validate:"required"`
		Gender   string `json:"gender" validate:"required"`
		Age      int    `json:"age" validate:"required,min=1"`
		Dob      string `json:"dob" validate:"required"`
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
		userDetails.Name,
		userDetails.Lastname,
		userDetails.Gender,
		userDetails.Age,
		userDetails.Dob,
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

// login by email and password
// func Login(c *fiber.Ctx) error {
// 	type LoginReq struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	var req LoginReq
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	// 1️⃣ Validate input
// 	if req.Email == "" || req.Password == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Email and password are required",
// 		})
// 	}

// 	// 2️⃣ Find user by email
// 	var user models.User
// 	if err := database.DB.
// 		Preload("Roles"). // Load roles if needed
// 		Where("email = ?", req.Email).
// 		First(&user).Error; err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Invalid email or password",
// 		})
// 	}

// 	// 3️⃣ Compare password with hashed password
// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Invalid email or password",
// 		})
// 	}

// 	// 4️⃣ Generate JWT token
// 	token, err := utils.GenerateJWT(user.ID, user.Email, user.FirebaseUID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to generate token",
// 		})
// 	}

// 	// 5️⃣ Return success response
// 	return c.Status(200).JSON(fiber.Map{
// 		"message": "Login successful",
// 		"token":   token,
// 		"user": fiber.Map{
// 			"id":       user.ID,
// 			"email":    user.Email,
// 			"username": user.Username,
// 			"uid":      user.FirebaseUID,
// 		},
// 	})
// }

// login by firebase token
func Loginn(c *fiber.Ctx) error {
	uid := c.Locals("uid").(string)

	var user models.User

	if err := database.DB.Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found, please register",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
	})
}

// get profile
func GetProfile(c *fiber.Ctx) error {
	uid := c.Locals("uid").(string)

	var user models.User
	if err := database.DB.Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}
