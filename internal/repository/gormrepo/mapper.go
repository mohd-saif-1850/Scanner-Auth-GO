package gormrepo

import (
	"scannerAuth/internal/domain"
	"time"
)

func toDomainUser(model *UserModel) *domain.User {
	if model == nil {
		return nil
	}

	return &domain.User{
		ID:                  model.ID,
		Email:               model.Email,
		PasswordHash:        model.PasswordHash,
		FullName:            model.FullName,
		Phone:               model.Phone,
		Role:                model.Role,
		IsActive:            model.IsActive,
		FailedLoginAttempts: model.FailedLoginAttempts,
		LockedUntil:         model.LockedUntil,
		LastLoginAt:         model.LastLoginAt,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
		DeletedAt:           deletedAt(model),
	}
}

func toUserModel(user *domain.User) *UserModel {
	if user == nil {
		return nil
	}

	return &UserModel{
		ID:                  user.ID,
		Email:               user.Email,
		PasswordHash:        user.PasswordHash,
		FullName:            user.FullName,
		Phone:               user.Phone,
		Role:                user.Role,
		IsActive:            user.IsActive,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastLoginAt:         user.LastLoginAt,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}

func deletedAt(model *UserModel) *time.Time {
	if model.DeletedAt.Valid {
		return &model.DeletedAt.Time
	}
	return nil
}

func toDomainToken(model *RefreshTokenModel) *domain.RefreshToken {
	if model == nil {
		return nil
	}
	return &domain.RefreshToken{
		ID:        model.ID,
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		FamilyID:  model.FamilyID,
		UserAgent: model.UserAgent,
		IPAddress: model.IPAddress,
		ExpiresAt: model.ExpiresAt,
		RevokedAt: model.RevokedAt,
		CreatedAt: model.CreatedAt,
	}
}

func toTokenModel(t *domain.RefreshToken) *RefreshTokenModel {
	if t == nil {
		return nil
	}
	return &RefreshTokenModel{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		FamilyID:  t.FamilyID,
		UserAgent: t.UserAgent,
		IPAddress: t.IPAddress,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		CreatedAt: t.CreatedAt,
	}
}
