package service

import (
	"context"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req dto.LoginRequest, meta dto.ClientMeta) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, id uuid.UUID, req dto.ChangePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID, password string) error
}
