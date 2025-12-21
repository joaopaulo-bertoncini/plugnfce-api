package main

import (
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	l := logger.NewZapLogger()
	l.Info("Starting ImobCheck migrations ...")
	dsn := "host=localhost user=plugnfce password=plugnfce dbname=plugnfce port=5432 sslmode=disable"
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Error("failed to connect to database: " + err.Error())
		panic(err)
	}

	err = runMigrations(db, "./migrations")
	if err != nil {
		l.Error("failed to run migrations: " + err.Error())
		return
	}
}

func runMigrations(db *gorm.DB, migrationsPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	// Convert to absolute path and ensure proper formatting
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	sourceURL := "file://" + absPath

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
