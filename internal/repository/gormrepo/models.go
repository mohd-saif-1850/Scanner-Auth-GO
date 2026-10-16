package gormrepo

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email               string         `gorm:"size:255;not null;uniqueIndex:uq_users_email_active,where:deleted_at IS NULL"`
	PasswordHash        string         `gorm:"not null"`
	FullName            string         `gorm:"size:120;not null"`
	Phone               *string        `gorm:"size:20"`
	Role                string         `gorm:"size:20;not null;default:user"`
	IsActive            bool           `gorm:"not null;default:true"`
	FailedLoginAttempts int            `gorm:"not null;default:0"`
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt     `gorm:"index"`
	RefreshTokens       []RefreshTokenModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (UserModel) TableName() string { return "users" }

type RefreshTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_rt_user_id"`
	TokenHash string     `gorm:"type:char(64);not null;uniqueIndex"`
	FamilyID  uuid.UUID  `gorm:"type:uuid;not null"`
	UserAgent *string
	IPAddress *string    `gorm:"type:inet"`
	ExpiresAt time.Time  `gorm:"not null;index:idx_rt_expires_at"`
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
