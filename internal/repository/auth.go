package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type IAuthRepository interface {
	CreateEmailCode(ctx context.Context, userId uuid.UUID, code string) error
	GetEmailCode(ctx context.Context, userId uuid.UUID) (model.CacheEmailCode, error)
	UpdateEmailCode(ctx context.Context, userId uuid.UUID, code string) error
	DeleteEmailCode(ctx context.Context, userId uuid.UUID) error
	GetExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) (*model.ExpiredTokenTime, error)
	UpdateExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) error
}

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateEmailCode(ctx context.Context, userId uuid.UUID, code string) error {
	cacheEmailCode := model.CacheEmailCode{
		UserId: userId,
		Code:   code,
	}
	return r.db.WithContext(ctx).Create(&cacheEmailCode).Error
}

func (r *AuthRepository) GetEmailCode(ctx context.Context, userId uuid.UUID) (model.CacheEmailCode, error) {
	var cacheEmailCode model.CacheEmailCode
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).First(&cacheEmailCode).Error; err != nil {
		return model.CacheEmailCode{}, err
	}
	return cacheEmailCode, nil
}

func (r *AuthRepository) UpdateEmailCode(ctx context.Context, userId uuid.UUID, code string) error {
	return r.db.WithContext(ctx).Model(&model.CacheEmailCode{}).Where("user_id = ?", userId).Update("code", code).Error
}

func (r *AuthRepository) DeleteEmailCode(ctx context.Context, userId uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("user_id = ?", userId).Delete(&model.CacheEmailCode{}).Error
}

func (r *AuthRepository) GetExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) (*model.ExpiredTokenTime, error) {
	var expiredTokenTime model.ExpiredTokenTime
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND subject_id = ?", organizationId, userId).First(&expiredTokenTime).Error; err != nil {
		return nil, err
	}
	return &expiredTokenTime, nil
}
func (r *AuthRepository) UpdateExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) error {
	// set expired time to now
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "subject_id"}, {Name: "organization_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"expired_time"}),
	}).Create(&model.ExpiredTokenTime{
		SubjectId:      userId,
		ExpiredTime:    time.Now(),
		OrganizationId: organizationId,
	}).Error
}
