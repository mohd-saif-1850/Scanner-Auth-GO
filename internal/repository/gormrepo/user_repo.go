package gormrepo

import (
	"context"

	"scannerAuth/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model UserModel

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return toDomain(&model), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return toDomain(&model), nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)

	return r.db.WithContext(ctx).
		Create(model).
		Error
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	model := toModel(user)

	return r.db.WithContext(ctx).
		Save(model).
		Error
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&UserModel{}, "id = ?", id).
		Error
}
