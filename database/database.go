package database

import (
	"Auth/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("POSTGRES_AUTHENTICATE")
	if dsn == "" {
		log.Fatal("❌ POSTGRES_AUTHENTICATE not set in .env")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Auto migrate models
	AutoMigrate()
}

func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("⚠️  Error getting database instance: %v", err)
			return
		}
		
		if err := sqlDB.Close(); err != nil {
			log.Printf("⚠️  Error closing database: %v", err)
		} else {
			log.Println("✅ Database connection closed")
		}
	}
}

func AutoMigrate() {
	// Add your models here
	DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.User_Details{},
	)
}













// package database

// import (
// 	"Auth/models"
// 	"fmt"
// 	"log"
// 	"os"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func Connect() {
// 	dsn := os.Getenv("POSTGRES_AUTHENTICATE")
// 	if dsn == "" {
// 		log.Fatal("❌ POSTGRES_AUTHENTICATE is not set")
// 	}

// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("❌ GORM open failed: %v", err)
// 	}

// 	// 🔥 IMPORTANT: real connection test
// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		log.Fatalf("❌ Failed to get sql.DB: %v", err)
// 	}

// 	if err := sqlDB.Ping(); err != nil {
// 		log.Fatalf("❌ Database ping failed: %v", err)
// 	}

// 	DB = db
// 	log.Println("✅ PostgreSQL connected successfully")
// 	// Auto migrate all models
// 	DB.AutoMigrate(&models.User{}, &models.User_Details{})
// 	fmt.Println("Database migrated success")
// }
