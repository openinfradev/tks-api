package repository

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IAlertRepository interface {
	Get(alertId uuid.UUID) (domain.Alert, error)
	GetByName(organizationId string, name string) (domain.Alert, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]domain.Alert, error)
	FetchPodRestart(organizationId string, start time.Time, end time.Time) ([]domain.Alert, error)
	Create(dto domain.Alert) (alertId uuid.UUID, err error)
	Update(dto domain.Alert) (err error)
	Delete(dto domain.Alert) (err error)

	CreateAlertAction(dto domain.AlertAction) (alertActionId uuid.UUID, err error)
}

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) IAlertRepository {
	return &AlertRepository{
		db: db,
	}
}

// Models
type Alert struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Name           string
	Code           string
	Description    string
	Grade          string
	Message        string
	ClusterId      domain.ClusterId
	Cluster        Cluster `gorm:"foreignKey:ClusterId"`
	Node           string
	CheckPoint     string
	GrafanaUrl     string
	Summary        string
	AlertActions   []AlertAction `gorm:"foreignKey:AlertId"`
	RawData        datatypes.JSON
	Status         domain.AlertActionStatus `gorm:"index"`
}

func (c *Alert) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

type AlertAction struct {
	gorm.Model

	ID      uuid.UUID `gorm:"primarykey"`
	AlertId uuid.UUID
	Content string
	Status  domain.AlertActionStatus
	TakerId *uuid.UUID `gorm:"type:uuid"`
	Taker   User       `gorm:"foreignKey:TakerId"`
}

func (c *AlertAction) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *AlertRepository) Get(alertId uuid.UUID) (out domain.Alert, err error) {
	var alert Alert
	res := r.db.Preload("AlertActions.Taker").Preload(clause.Associations).First(&alert, "id = ?", alertId)
	if res.Error != nil {
		return domain.Alert{}, res.Error
	}
	out = reflectAlert(alert)
	return
}

func (r *AlertRepository) GetByName(organizationId string, name string) (out domain.Alert, err error) {
	var alert Alert
	res := r.db.Preload("AlertActions.Taker").Preload(clause.Associations).First(&alert, "organization_id = ? AND name = ?", organizationId, name)

	if res.Error != nil {
		return domain.Alert{}, res.Error
	}
	out = reflectAlert(alert)
	return
}

func (r *AlertRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []domain.Alert, err error) {
	var alerts []Alert
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}

	filterFunc := CombinedGormFilter("alerts", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&Alert{}).
		Preload("AlertActions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Preload("AlertActions.Taker").
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Joins("join clusters on clusters.id = alerts.cluster_id AND clusters.status = 2").
		Where("alerts.organization_id = ?", organizationId))

	db.Count(&pg.TotalRows)

	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))
	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&alerts)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, alert := range alerts {
		out = append(out, reflectAlert(alert))
	}
	return
}

func (r *AlertRepository) FetchPodRestart(organizationId string, start time.Time, end time.Time) (out []domain.Alert, err error) {
	var alerts []Alert
	res := r.db.Preload(clause.Associations).Order("created_at DESC").
		Where("organization_id = ? AND name = 'pod-restart-frequently' AND created_at BETWEEN ? AND ?", organizationId, start, end).
		Find(&alerts)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, alert := range alerts {
		out = append(out, reflectAlert(alert))
	}
	return
}

func (r *AlertRepository) Create(dto domain.Alert) (alertId uuid.UUID, err error) {
	alert := Alert{
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

func (r *AlertRepository) Update(dto domain.Alert) (err error) {
	res := r.db.Model(&Alert{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) Delete(dto domain.Alert) (err error) {
	res := r.db.Delete(&Alert{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertRepository) CreateAlertAction(dto domain.AlertAction) (alertActionId uuid.UUID, err error) {
	alert := AlertAction{
		AlertId: dto.AlertId,
		Content: dto.Content,
		Status:  dto.Status,
		TakerId: dto.TakerId,
	}
	res := r.db.Create(&alert)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	res = r.db.Model(&Alert{}).
		Where("id = ?", dto.AlertId).
		Update("status", dto.Status)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return alert.ID, nil
}

func reflectAlert(alert Alert) (out domain.Alert) {
	if err := serializer.Map(alert, &out); err != nil {
		log.Error(err)
	}

	outAlertActions := make([]domain.AlertAction, len(alert.AlertActions))
	if err := serializer.Map(outAlertActions, &out.AlertActions); err != nil {
		log.Error(err)
	}
	return
}
