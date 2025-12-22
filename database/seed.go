// database/seed.go
package database

import "Auth/models"

func SeedRoles() {
	roles := []string{"user", "admin"}

	for _, r := range roles {
		var role models.Role
		if err := DB.Where("name = ?", r).First(&role).Error; err != nil {
			DB.Create(&models.Role{Name: r})
		}
	}
}
