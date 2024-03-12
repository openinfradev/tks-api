package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAuthRepository interface {
	CreateEmailCode(ctx context.Context, userId uuid.UUID, code string) error
	GetEmailCode(ctx context.Context, userId uuid.UUID) (CacheEmailCode, error)
	UpdateEmailCode(ctx context.Context, userId uuid.UUID, code string) error
	DeleteEmailCode(ctx context.Context, userId uuid.UUID) error
}

type AuthRepository struct {
	db *gorm.DB
}

// Models

type CacheEmailCode struct {
	gorm.Model

	UserId uuid.UUID `gorm:"not null"`
	Code   string    `gorm:"type:varchar(6);not null"`
}

func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateEmailCode(ctx context.Context, userId uuid.UUID, code string) error {
	cacheEmailCode := CacheEmailCode{
		UserId: userId,
		Code:   code,
	}
	return r.db.WithContext(ctx).Create(&cacheEmailCode).Error
}

func (r *AuthRepository) GetEmailCode(ctx context.Context, userId uuid.UUID) (CacheEmailCode, error) {
	var cacheEmailCode CacheEmailCode
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).First(&cacheEmailCode).Error; err != nil {
		return CacheEmailCode{}, err
	}
	return cacheEmailCode, nil
}

func (r *AuthRepository) UpdateEmailCode(ctx context.Context, userId uuid.UUID, code string) error {
	return r.db.WithContext(ctx).Model(&CacheEmailCode{}).Where("user_id = ?", userId).Update("code", code).Error
}

func (r *AuthRepository) DeleteEmailCode(ctx context.Context, userId uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("user_id = ?", userId).Delete(&CacheEmailCode{}).Error
}
