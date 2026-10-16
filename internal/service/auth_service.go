package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"scannerAuth/internal/domain"
	"scannerAuth/internal/dto"
	"scannerAuth/internal/hasher"
	"scannerAuth/internal/pkg/token"
	"scannerAuth/internal/repository"

	"github.com/google/uuid"
)

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	txManager repository.TxManager
	hasher    hasher.PasswordHasher
	tokenMgr  token.Manager
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	txManager repository.TxManager,
	hasher hasher.PasswordHasher,
	tokenMgr token.Manager,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		txManager: txManager,
		hasher:    hasher,
		tokenMgr:  tokenMgr,
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
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
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
		Role:         "user",
		IsActive:     true,
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

func (s *authService) Login(ctx context.Context, req dto.LoginRequest, meta dto.ClientMeta) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Return same error for wrong email and wrong password (security)
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive || user.DeletedAt != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, domain.ErrAccountLocked
	}

	if err := s.hasher.Compare(ctx, user.PasswordHash, req.Password); err != nil {
		_ = s.userRepo.RecordLoginAttempt(ctx, user.ID, false)
		return nil, domain.ErrInvalidCredentials
	}

	_ = s.userRepo.RecordLoginAttempt(ctx, user.ID, true)

	accessToken, expiresAt, err := s.tokenMgr.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, hashRefresh, err := s.tokenMgr.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashRefresh,
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if meta.UserAgent != "" {
		rt.UserAgent = &meta.UserAgent
	}
	if meta.IPAddress != "" {
		rt.IPAddress = &meta.IPAddress
	}

	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return nil, err
	}

	return buildAuthResponse(accessToken, rawRefresh, expiresAt, user), nil
}

func (s *authService) Refresh(ctx context.Context, rawToken string) (*dto.AuthResponse, error) {
	hash := s.tokenMgr.HashRefreshToken(rawToken)

	rt, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now()) {
		// Token reuse detected — revoke entire family
		_ = s.tokenRepo.RevokeFamily(ctx, rt.FamilyID)
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive || user.DeletedAt != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Rotate: revoke old, issue new
	var (
		accessToken  string
		expiresAt    time.Time
		rawRefresh   string
		hashRefresh  string
	)

	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.tokenRepo.Revoke(txCtx, rt.ID); err != nil {
			return err
		}

		var terr error
		accessToken, expiresAt, terr = s.tokenMgr.GenerateAccessToken(user)
		if terr != nil {
			return terr
		}

		rawRefresh, hashRefresh, terr = s.tokenMgr.GenerateRefreshToken()
		if terr != nil {
			return terr
		}

		newRT := &domain.RefreshToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: hashRefresh,
			FamilyID:  rt.FamilyID, // same family
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		return s.tokenRepo.Save(txCtx, newRT)
	})
	if err != nil {
		return nil, err
	}

	return buildAuthResponse(accessToken, rawRefresh, expiresAt, user), nil
}

func (s *authService) Logout(ctx context.Context, rawToken string) error {
	hash := s.tokenMgr.HashRefreshToken(rawToken)
	rt, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil // already gone — treat as success
	}
	return s.tokenRepo.Revoke(ctx, rt.ID)
}

// helpers

func validPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}
	hasLetter, hasDigit := false, false
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

func buildAuthResponse(accessToken, refreshToken string, expiresAt time.Time, user *domain.User) *dto.AuthResponse {
	ur := domainUserToResponse(user)
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Until(expiresAt).Milliseconds() / 1000,
		User:         ur,
	}
}

// DomainUserToResponse is exported so handlers can build UserResponse from domain.User.
func DomainUserToResponse(u *domain.User) dto.UserResponse {
	return domainUserToResponse(u)
}

func domainUserToResponse(u *domain.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:       u.ID.String(),
		Email:    u.Email,
		FullName: u.FullName,
		Phone:    u.Phone,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &s
	}
	resp.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	resp.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	return resp
}
