package firebase

import (
	"Auth/models"
)

// SetUserClaims sets custom claims for a Firebase user
// Helper function to generate Firebase custom claims
// Extract role names
func GetRoleNames(roles []models.Role) []string {
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return roleNames
}

// Generate Firebase custom claims
func GenerateUserClaims(user models.User) map[string]interface{} {
	claims := make(map[string]interface{})
	claims["user_id"] = user.ID
 
	var roleNames []string
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
		claims[r.Name] = true
	}
	claims["roles"] = roleNames

	return claims
}
