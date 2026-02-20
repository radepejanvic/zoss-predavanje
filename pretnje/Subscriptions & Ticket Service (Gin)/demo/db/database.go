package db

import (
	"fmt"
	"gin-subscription-service/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	var err error
	var dialector gorm.Dialector

	// Check if running in Docker with PostgreSQL
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		// PostgreSQL configuration from environment
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dbPort := os.Getenv("DB_PORT")

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			dbHost, dbUser, dbPassword, dbName, dbPort)

		dialector = postgres.Open(dsn)
		log.Println("Using PostgreSQL database")
	} else {
		// Default to SQLite for local development
		dialector = sqlite.Open("subscriptions.db")
		log.Println("Using SQLite database")
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&models.Subscription{}, &models.Ticket{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully")
}
