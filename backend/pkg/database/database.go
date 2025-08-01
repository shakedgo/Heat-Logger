package database

import (
	"log"

	"heat-logger/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() error {
	var err error

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(&models.DailyRecord{})
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
