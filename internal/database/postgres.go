package database

import (
	"fmt"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect initializes the GORM PostgreSQL connection and auto-migrates all models.
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate creates/updates tables to match domain structs.
	// In production you'd use a proper migration tool (e.g., golang-migrate).
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Post{},
		&domain.Image{},
		&domain.RefreshToken{},
	); err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	return db, nil
}
