package gormrepo

import (
	"context"
	"time"

	"scannerAuth/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, u *domain.User) error {
	model := toUserModel(u)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return mapError(err)
	}
	u.ID = model.ID
	u.CreatedAt = model.CreatedAt
	u.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, mapError(err)
	}
	return toDomainUser(&model), nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		return nil, mapError(err)
	}
	return toDomainUser(&model), nil
}

func (r *GormUserRepository) Update(ctx context.Context, u *domain.User) error {
	model := toUserModel(u)
	err := r.db.WithContext(ctx).
		Select("email", "full_name", "phone", "updated_at").
		Updates(model).Error
	return mapError(err)
}

func (r *GormUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password_hash": hash,
			"updated_at":    time.Now(),
		}).Error
	return mapError(err)
}

func (r *GormUserRepository) RecordLoginAttempt(ctx context.Context, id uuid.UUID, success bool) error {
	if success {
		err := r.db.WithContext(ctx).
			Model(&UserModel{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"failed_login_attempts": 0,
				"locked_until":          nil,
				"last_login_at":         time.Now(),
				"updated_at":            time.Now(),
			}).Error
		return mapError(err)
	}

	// On failure: increment counter; lock after 5 attempts for 15 minutes
	err := r.db.WithContext(ctx).Exec(`
		UPDATE users SET
			failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE WHEN failed_login_attempts + 1 >= 5
				THEN NOW() + INTERVAL '15 minutes'
				ELSE locked_until
			END,
			updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id).Error
	return mapError(err)
}

func (r *GormUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&UserModel{}).Error
	return mapError(err)
}
