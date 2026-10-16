package gormrepo

import (
	"scannerAuth/internal/domain"
	"time"
)

func toDomain(model *UserModel) *domain.User {
	if model == nil {
		return nil
	}

	return &domain.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FullName:     model.FullName,
		Phone:        model.Phone,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		DeletedAt:    deletedAt(model),
	}
}

func toModel(user *domain.User) *UserModel {
	if user == nil {
		return nil
	}

	return &UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		Phone:        user.Phone,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func deletedAt(model *UserModel) *time.Time {
	if model.DeletedAt.Valid {
		return &model.DeletedAt.Time
	}

	return nil
}
