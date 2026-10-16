package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"scannerAuth/internal/config"
)

func Connect(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.GinMode != "release" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{
		Logger:          gormLogger,
		TranslateError:  true,
		PrepareStmt:     true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxConns)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	slog.New(slog.NewJSONHandler(os.Stdout, nil)).Info("database connected")

	return db, sqlDB, nil
}
