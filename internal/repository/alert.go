package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type IAlertRepository interface {
	Get(alertId uuid.UUID) (domain.Alert, error)
	GetByName(organizationId string, name string) (domain.Alert, error)
	Fetch(organizationId string) ([]domain.Alert, error)
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

func (r *AlertRepository) Fetch(organizationId string) (out []domain.Alert, err error) {
	var alerts []Alert
	res := r.db.Preload("AlertActions", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Preload("AlertActions.Taker").
		Preload("Cluster", "status = 2").
		Preload("Organization").
		Order("created_at desc").Find(&alerts, "organization_id = ?", organizationId)
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
		Where("organization_id = ? AND created_at BETWEEN ? AND ?", organizationId, start, end).
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
	return alert.ID, nil
}

func reflectAlert(alert Alert) domain.Alert {
	outAlertActions := make([]domain.AlertAction, len(alert.AlertActions))
	for i, alertAction := range alert.AlertActions {
		outAlertActions[i] = reflectAlertAction(alertAction)
	}

	return domain.Alert{
		ID:             alert.ID,
		OrganizationId: alert.OrganizationId,
		Name:           alert.Name,
		Description:    alert.Description,
		Message:        alert.Message,
		Code:           alert.Code,
		Grade:          alert.Grade,
		ClusterId:      alert.ClusterId,
		Cluster:        reflectSimpleCluster(alert.Cluster),
		GrafanaUrl:     alert.GrafanaUrl,
		Node:           alert.Node,
		CheckPoint:     alert.CheckPoint,
		Summary:        alert.Summary,
		AlertActions:   outAlertActions,
		RawData:        alert.RawData,
		CreatedAt:      alert.CreatedAt,
		UpdatedAt:      alert.UpdatedAt,
	}
}

func reflectAlertAction(alertAction AlertAction) domain.AlertAction {
	return domain.AlertAction{
		ID:        alertAction.ID,
		AlertId:   alertAction.AlertId,
		Content:   alertAction.Content,
		Status:    alertAction.Status,
		TakerId:   alertAction.TakerId,
		Taker:     reflectSimpleUser(alertAction.Taker),
		CreatedAt: alertAction.CreatedAt,
		UpdatedAt: alertAction.UpdatedAt,
	}
}
