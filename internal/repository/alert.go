package repository

import (
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
	Get(alertId uuid.UUID) (model.Alert, error)
	GetByName(organizationId string, name string) (model.Alert, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]model.Alert, error)
	FetchPodRestart(organizationId string, start time.Time, end time.Time) ([]model.Alert, error)
	Create(dto model.Alert) (alertId uuid.UUID, err error)
	Update(dto model.Alert) (err error)
	Delete(dto model.Alert) (err error)

	CreateAlertAction(dto model.AlertAction) (alertActionId uuid.UUID, err error)
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
func (r *AlertRepository) Get(alertId uuid.UUID) (out model.Alert, err error) {
	res := r.db.Preload("AlertActions.Taker").Preload(clause.Associations).First(&out, "id = ?", alertId)
	if res.Error != nil {
		return model.Alert{}, res.Error
	}
	return
}

func (r *AlertRepository) GetByName(organizationId string, name string) (out model.Alert, err error) {
	res := r.db.Preload("AlertActions.Taker").Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.Alert{}, res.Error
	}
	return
}

func (r *AlertRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []model.Alert, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.Model(&model.Alert{}).
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

func (r *AlertRepository) FetchPodRestart(organizationId string, start time.Time, end time.Time) (out []model.Alert, err error) {
	res := r.db.Preload(clause.Associations).Order("created_at DESC").
		Where("organization_id = ? AND name = 'pod-restart-frequently' AND created_at BETWEEN ? AND ?", organizationId, start, end).
		Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AlertRepository) Create(dto model.Alert) (alertId uuid.UUID, err error) {
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
	res := r.db.Create(&alert)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return alert.ID, nil
}

func (r *AlertRepository) Update(dto model.Alert) (err error) {
	res := r.db.Model(&model.Alert{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) Delete(dto model.Alert) (err error) {
	res := r.db.Delete(&model.Alert{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) CreateAlertAction(dto model.AlertAction) (alertActionId uuid.UUID, err error) {
	alert := model.AlertAction{
		AlertId: dto.AlertId,
		Content: dto.Content,
		Status:  dto.Status,
		TakerId: dto.TakerId,
	}
	res := r.db.Create(&alert)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	res = r.db.Model(&model.Alert{}).
		Where("id = ?", dto.AlertId).
		Update("status", dto.Status)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return alert.ID, nil
}
