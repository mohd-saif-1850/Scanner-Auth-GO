package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/hasher"
	"scannerAuth/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authService struct {
	userRepo repository.UserRepository
	hasher   hasher.PasswordHasher
}

func NewAuthService(
	userRepo repository.UserRepository,
	hasher hasher.PasswordHasher,
) AuthService {
	return &authService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)

	if !validPassword(req.Password) {
		return nil, errors.New("password must be 8-72 characters and contain at least one letter and one digit")
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, email)

	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailTaken
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := s.hasher.Hash(ctx, req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
	}

	if strings.TrimSpace(req.Phone) != "" {
		phone := strings.TrimSpace(req.Phone)
		user.Phone = &phone
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func validPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}

		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}
