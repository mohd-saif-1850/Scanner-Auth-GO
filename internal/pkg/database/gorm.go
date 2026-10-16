package database

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"scannerAuth/internal/config"
)

func Connect(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxConns)

	return db, sqlDB, nil
}
