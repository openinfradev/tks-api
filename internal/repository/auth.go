package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAuthRepository interface {
	CreateEmailCode(userId uuid.UUID, code string) error
	GetEmailCode(userId uuid.UUID) (CacheEmailCode, error)
	UpdateEmailCode(userId uuid.UUID, code string) error
	DeleteEmailCode(userId uuid.UUID) error
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

func (r *AuthRepository) CreateEmailCode(userId uuid.UUID, code string) error {
	cacheEmailCode := CacheEmailCode{
		UserId: userId,
		Code:   code,
	}
	return r.db.Create(&cacheEmailCode).Error
}

func (r *AuthRepository) GetEmailCode(userId uuid.UUID) (CacheEmailCode, error) {
	var cacheEmailCode CacheEmailCode
	if err := r.db.Where("user_id = ?", userId).First(&cacheEmailCode).Error; err != nil {
		return CacheEmailCode{}, err
	}
	return cacheEmailCode, nil
}

func (r *AuthRepository) UpdateEmailCode(userId uuid.UUID, code string) error {
	return r.db.Model(&CacheEmailCode{}).Where("user_id = ?", userId).Update("code", code).Error
}

func (r *AuthRepository) DeleteEmailCode(userId uuid.UUID) error {
	return r.db.Unscoped().Where("user_id = ?", userId).Delete(&CacheEmailCode{}).Error
}
