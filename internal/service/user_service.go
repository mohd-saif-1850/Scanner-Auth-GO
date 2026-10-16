package service

import (
	"context"
	"errors"
	"strings"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/hasher"
	"scannerAuth/internal/repository"

	"github.com/google/uuid"
)

type userService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	txManager repository.TxManager
	hasher    hasher.PasswordHasher
}

func NewUserService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	txManager repository.TxManager,
	hasher hasher.PasswordHasher,
) UserService {
	return &userService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		txManager: txManager,
		hasher:    hasher,
	}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := domainUserToResponse(user)
	return &resp, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply only provided fields
	if req.FullName != "" {
		user.FullName = strings.TrimSpace(req.FullName)
	}
	if req.Phone != "" {
		phone := strings.TrimSpace(req.Phone)
		user.Phone = &phone
	}
	if req.Email != "" {
		newEmail := strings.ToLower(strings.TrimSpace(req.Email))
		if newEmail != user.Email {
			// Check uniqueness
			existing, err := s.userRepo.GetByEmail(ctx, newEmail)
			if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
				return nil, err
			}
			if existing != nil {
				return nil, domain.ErrEmailTaken
			}
			user.Email = newEmail
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := domainUserToResponse(user)
	return &resp, nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.hasher.Compare(ctx, user.PasswordHash, req.CurrentPassword); err != nil {
		return domain.ErrInvalidCredentials
	}

	if req.CurrentPassword == req.NewPassword {
		return errors.New("new password must differ from current password")
	}

	newHash, err := s.hasher.Hash(ctx, req.NewPassword)
	if err != nil {
		return err
	}

	return s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.UpdatePassword(txCtx, id, newHash); err != nil {
			return err
		}
		return s.tokenRepo.RevokeAllForUser(txCtx, id)
	})
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID, password string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.hasher.Compare(ctx, user.PasswordHash, password); err != nil {
		return domain.ErrInvalidCredentials
	}

	return s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.SoftDelete(txCtx, id); err != nil {
			return err
		}
		return s.tokenRepo.RevokeAllForUser(txCtx, id)
	})
}
