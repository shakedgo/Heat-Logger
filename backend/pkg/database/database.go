package database

import (
	"heat-logger/internal/config"
	"log"

	"heat-logger/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection and runs migrations
func InitDatabase(cfg *config.Config) error {
	var err error

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
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

	// Migrate existing records to have 'global' as default UserID
	err = migrateExistingRecords()
	if err != nil {
		log.Printf("Warning: Failed to migrate existing records: %v", err)
	}

	log.Printf("Database initialized successfully at %s", cfg.Database.Path)
	return nil
}

// migrateExistingRecords updates existing records without UserID to use 'global'
func migrateExistingRecords() error {
	// Update any records that have empty or null UserID to 'global'
	result := DB.Model(&models.DailyRecord{}).Where("user_id = '' OR user_id IS NULL").Update("user_id", "global")
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("Migrated %d existing records to use 'global' UserID", result.RowsAffected)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
