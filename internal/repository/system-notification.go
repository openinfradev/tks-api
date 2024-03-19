package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type ISystemNotificationRepository interface {
	Get(ctx context.Context, systemNotificationId uuid.UUID) (model.SystemNotification, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.SystemNotification, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotification, error)
	FetchPodRestart(ctx context.Context, organizationId string, start time.Time, end time.Time) ([]model.SystemNotification, error)
	Create(ctx context.Context, dto model.SystemNotification) (systemNotificationId uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotification) (err error)
	Delete(ctx context.Context, dto model.SystemNotification) (err error)

	CreateSystemNotificationAction(ctx context.Context, dto model.SystemNotificationAction) (systemNotificationActionId uuid.UUID, err error)
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

func (r *SystemNotificationRepository) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.SystemNotification, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Preload("SystemNotificationActions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Preload("SystemNotificationActions.Taker").
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Joins("join clusters on clusters.id = system_notifications.cluster_id AND clusters.status = 2").
		Where("system_notifications.organization_id = ?", organizationId), &out)

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
	systemNotification := model.SystemNotification{
		ID:             uuid.New(),
		OrganizationId: dto.OrganizationId,
		Name:           dto.Name,
		Code:           dto.Code,
		Message:        dto.Message,
		Description:    dto.Description,
		Grade:          dto.Grade,
		ClusterId:      dto.ClusterId,
		Node:           dto.Node,
		GrafanaUrl:     dto.GrafanaUrl,
		CheckPoint:     dto.CheckPoint,
		Summary:        dto.Summary,
		RawData:        dto.RawData,
		Status:         domain.SystemNotificationActionStatus_CREATED,
	}
	res := r.db.WithContext(ctx).Create(&systemNotification)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return systemNotification.ID, nil
}

func (r *SystemNotificationRepository) Update(ctx context.Context, dto model.SystemNotification) (err error) {
	res := r.db.WithContext(ctx).Model(&model.SystemNotification{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description})
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
