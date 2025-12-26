package models

import (
	"time"

	"gorm.io/gorm"
)

// Status constants
const (
	StatusInactive = 0
	StatusActive   = 1
)

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
type JobType struct {
	gorm.Model
	Name   string `json:"name" gorm:"unique;not null"`
	Status int    `json:"status"`
	Jobs   []Job  `gorm:"foreignKey:JobTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type Company struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null"`
	Email       string `json:"email" gorm:"unique;not null"`
	Address     string `json:"address" gorm:"not null"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Status      int    `json:"status"`
	Jobs        []Job  `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Job struct {
	gorm.Model
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	SalaryStart int64     `json:"salary_start"`
	SalaryEnd   int64     `json:"salary_end"`
	Type        string    `json:"type" gorm:"not null;index"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      int       `json:"status"`
	JobTypeID   uint      `gorm:"not null;index"`
	CompanyID   uint      `gorm:"not null;index"`

	JobType JobType `json:"job_type"`
	Company Company `json:"company"`
}
