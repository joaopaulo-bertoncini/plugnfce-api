package database

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(ctx context.Context, databaseURL string, env string) error {
	var err error

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if env == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Connect to database
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
