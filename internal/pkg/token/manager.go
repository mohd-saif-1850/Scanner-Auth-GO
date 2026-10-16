package token

import (
	"scannerAuth/internal/domain"
	"time"
)

type Manager interface {
	GenerateAccessToken(u *domain.User) (string, time.Time, error)
	ValidateAccessToken(token string) (*domain.Claims, error)
	GenerateRefreshToken() (raw string, hash string, err error)
	HashRefreshToken(raw string) string
}
