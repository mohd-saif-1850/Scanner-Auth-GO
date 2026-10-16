package service

import (
	"context"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error)
}
