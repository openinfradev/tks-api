package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

// Interfaces
type ISystemNotificationRepository interface {
	Get(ctx context.Context, systemNotificationId uuid.UUID) (model.SystemNotification, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.SystemNotification, error)
	FetchSystemNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotification, error)
	FetchPolicyNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotification, error)
	FetchPodRestart(ctx context.Context, organizationId string, start time.Time, end time.Time) ([]model.SystemNotification, error)
	Create(ctx context.Context, dto model.SystemNotification) (systemNotificationId uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotification) (err error)
	Delete(ctx context.Context, dto model.SystemNotification) (err error)
	CreateSystemNotificationAction(ctx context.Context, dto model.SystemNotificationAction) (systemNotificationActionId uuid.UUID, err error)
	UpdateRead(ctx context.Context, systemNotificationId uuid.UUID, user model.User) (err error)
}

type SystemNotificationRepository struct {
	db *gorm.DB
}

func NewSystemNotificationRepository(db *gorm.DB) ISystemNotificationRepository {
	return &SystemNotificationRepository{
		db: db,
	}
}

// Logics
func (r *SystemNotificationRepository) Get(ctx context.Context, systemNotificationId uuid.UUID) (out model.SystemNotification, err error) {
	res := r.db.WithContext(ctx).Preload("SystemNotificationActions.Taker").Preload(clause.Associations).First(&out, "id = ?", systemNotificationId)
	if res.Error != nil {
		return model.SystemNotification{}, res.Error
	}

	return
}

func (r *SystemNotificationRepository) GetByName(ctx context.Context, organizationId string, name string) (out model.SystemNotification, err error) {
	res := r.db.WithContext(ctx).Preload("SystemNotificationActions.Taker").Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.SystemNotification{}, res.Error
	}
	return
}

func (r *SystemNotificationRepository) FetchSystemNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.SystemNotification, err error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return out, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	db := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Preload("SystemNotificationActions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Preload("SystemNotificationActions.Taker").
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Joins("join clusters on clusters.id = system_notifications.cluster_id AND clusters.status = 2").
		Joins("left outer join system_notification_rules ON system_notification_rules.id = system_notifications.system_notification_rule_id").
		Joins("left outer join system_notification_rule_users ON system_notification_rule_users.system_notification_rule_id = system_notifications.system_notification_rule_id").
		Where("system_notification_rule_users.user_id is null OR system_notification_rule_users.user_id = ?", userInfo.GetUserId()).
		Where("system_notifications.organization_id = ? AND system_notifications.notification_type = 'SYSTEM_NOTIFICATION'", organizationId)

	readFilter := pg.GetFilter("read")
	if readFilter != nil {
		if readFilter.Values[0] == "true" {
			db.Joins("join system_notification_users on system_notification_users.system_notification_id = system_notifications.id AND system_notification_users.user_id = ?", userInfo.GetUserId())
		} else {
			db.Joins("left outer join system_notification_users on system_notification_users.system_notification_id = system_notifications.id AND system_notification_users.user_id = ?", userInfo.GetUserId()).
				Where("system_notification_users.user_id is null")
		}
	}

	_, res := pg.Fetch(db, &out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationRepository) FetchPolicyNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.SystemNotification, err error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return out, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	db := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Joins("join clusters on clusters.id = system_notifications.cluster_id AND clusters.status = 2").
		Where("system_notifications.organization_id = ? AND system_notifications.notification_type = 'POLICY_NOTIFICATION'", organizationId)

	readFilter := pg.GetFilter("read")
	if readFilter != nil {
		if readFilter.Values[0] == "true" {
			db.Joins("join system_notification_users on system_notification_users.system_notification_id = system_notifications.id AND system_notification_users.user_id = ?", userInfo.GetUserId())
		} else {
			db.Joins("left outer join system_notification_users on system_notification_users.system_notification_id = system_notifications.id AND system_notification_users.user_id = ?", userInfo.GetUserId()).
				Where("system_notification_users.user_id is null")
		}
	}

	_, res := pg.Fetch(db, &out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationRepository) FetchPodRestart(ctx context.Context, organizationId string, start time.Time, end time.Time) (out []model.SystemNotification, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).Order("created_at DESC").
		Where("organization_id = ? AND name = 'pod-restart-frequently' AND created_at BETWEEN ? AND ?", organizationId, start, end).
		Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationRepository) Create(ctx context.Context, dto model.SystemNotification) (systemNotificationId uuid.UUID, err error) {

	dto.ID = uuid.New()
	dto.Status = domain.SystemNotificationActionStatus_CREATED
	res := r.db.WithContext(ctx).Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return dto.ID, nil
}

func (r *SystemNotificationRepository) Update(ctx context.Context, dto model.SystemNotification) (err error) {
	res := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"MessageTitle": dto.MessageTitle,
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *SystemNotificationRepository) Delete(ctx context.Context, dto model.SystemNotification) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.SystemNotification{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *SystemNotificationRepository) CreateSystemNotificationAction(ctx context.Context, dto model.SystemNotificationAction) (systemNotificationActionId uuid.UUID, err error) {
	systemNotification := model.SystemNotificationAction{
		ID:                   uuid.New(),
		SystemNotificationId: dto.SystemNotificationId,
		Content:              dto.Content,
		Status:               dto.Status,
		TakerId:              dto.TakerId,
	}
	res := r.db.WithContext(ctx).Create(&systemNotification)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	res = r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("id = ?", dto.SystemNotificationId).
		Update("status", dto.Status)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return systemNotification.ID, nil
}

func (r *SystemNotificationRepository) UpdateRead(ctx context.Context, systemNotificationId uuid.UUID, user model.User) (err error) {
	var systemNotification = model.SystemNotification{}
	res := r.db.WithContext(ctx).First(&systemNotification, "id = ?", systemNotificationId)
	if res.Error != nil {
		return res.Error
	}

	users := []model.User{user}
	err = r.db.WithContext(ctx).Model(&systemNotification).Association("Readers").Append(users)
	if err != nil {
		return err
	}
	return nil
}
