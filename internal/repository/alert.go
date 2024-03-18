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
type IAlertRepository interface {
	Get(ctx context.Context, alertId uuid.UUID) (model.Alert, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.Alert, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.Alert, error)
	FetchPodRestart(ctx context.Context, organizationId string, start time.Time, end time.Time) ([]model.Alert, error)
	Create(ctx context.Context, dto model.Alert) (alertId uuid.UUID, err error)
	Update(ctx context.Context, dto model.Alert) (err error)
	Delete(ctx context.Context, dto model.Alert) (err error)

	CreateAlertAction(ctx context.Context, dto model.AlertAction) (alertActionId uuid.UUID, err error)
}

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) IAlertRepository {
	return &AlertRepository{
		db: db,
	}
}

// Logics
func (r *AlertRepository) Get(ctx context.Context, alertId uuid.UUID) (out model.Alert, err error) {
	res := r.db.WithContext(ctx).Preload("AlertActions.Taker").Preload(clause.Associations).First(&out, "id = ?", alertId)
	if res.Error != nil {
		return model.Alert{}, res.Error
	}
	return
}

func (r *AlertRepository) GetByName(ctx context.Context, organizationId string, name string) (out model.Alert, err error) {
	res := r.db.WithContext(ctx).Preload("AlertActions.Taker").Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.Alert{}, res.Error
	}
	return
}

func (r *AlertRepository) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.Alert, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.Alert{}).
		Preload("AlertActions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Preload("AlertActions.Taker").
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Joins("join clusters on clusters.id = alerts.cluster_id AND clusters.status = 2").
		Where("alerts.organization_id = ?", organizationId), &out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AlertRepository) FetchPodRestart(ctx context.Context, organizationId string, start time.Time, end time.Time) (out []model.Alert, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).Order("created_at DESC").
		Where("organization_id = ? AND name = 'pod-restart-frequently' AND created_at BETWEEN ? AND ?", organizationId, start, end).
		Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AlertRepository) Create(ctx context.Context, dto model.Alert) (alertId uuid.UUID, err error) {
	alert := model.Alert{
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
		Status:         domain.AlertActionStatus_CREATED,
	}
	res := r.db.WithContext(ctx).Create(&alert)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return alert.ID, nil
}

func (r *AlertRepository) Update(ctx context.Context, dto model.Alert) (err error) {
	res := r.db.WithContext(ctx).Model(&model.Alert{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) Delete(ctx context.Context, dto model.Alert) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.Alert{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) CreateAlertAction(ctx context.Context, dto model.AlertAction) (alertActionId uuid.UUID, err error) {
	alert := model.AlertAction{
		AlertId: dto.AlertId,
		Content: dto.Content,
		Status:  dto.Status,
		TakerId: dto.TakerId,
	}
	res := r.db.WithContext(ctx).Create(&alert)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	res = r.db.WithContext(ctx).Model(&model.Alert{}).
		Where("id = ?", dto.AlertId).
		Update("status", dto.Status)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return alert.ID, nil
}
