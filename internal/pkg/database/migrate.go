package database

import (
	"fmt"

	"gorm.io/gorm"

	"scannerAuth/internal/repository/gormrepo"
)

func Migrate(db *gorm.DB) error {
	// Enable pgcrypto for gen_random_uuid()
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}

	if err := db.AutoMigrate(
		&gormrepo.UserModel{},
		&gormrepo.RefreshTokenModel{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
