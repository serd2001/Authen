package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey"`
	FirebaseUID  string `gorm:"uniqueIndex;not null"`
	Username     string `json:"username" gorm:"uniqueIndex;not null"`
	Email        string `json:"email" gorm:"uniqueIndex"`
	Password     string `json:"password"`
	Provider     string `json:"provider"` // password, google, etc.
	User_Details User_Details
	Roles        []Role `gorm:"many2many:user_roles;"` // many to many
}
type User_Details struct {
	gorm.Model
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
	Dob      string `json:"dob"`
	UserID   uint   `gorm:"unique"`
}
type Role struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Users []User `gorm:"many2many:user_roles;"` //  Many-to-Many
}
