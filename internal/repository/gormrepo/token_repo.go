package gormrepo

import (
	"context"
	"time"

	"scannerAuth/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormTokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *GormTokenRepository {
	return &GormTokenRepository{db: db}
}

func (r *GormTokenRepository) Save(ctx context.Context, t *domain.RefreshToken) error {
	model := toTokenModel(t)
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return mapError(err)
	}
	t.ID = model.ID
	t.CreatedAt = model.CreatedAt
	return nil
}

func (r *GormTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&model).Error
	if err != nil {
		return nil, mapError(err)
	}
	return toDomainToken(&model), nil
}

func (r *GormTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
	return mapError(err)
}

func (r *GormTokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", now).Error
	return mapError(err)
}

func (r *GormTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
	return mapError(err)
}
