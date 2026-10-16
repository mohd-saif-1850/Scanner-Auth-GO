package gormrepo

import (
	"errors"

	"scannerAuth/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const pgUniqueViolation = "23505"

func mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgUniqueViolation {
			return domain.ErrEmailTaken
		}
	}

	return err
}
