package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IAppGroupRepository interface {
	Fetch(clusterId domain.ClusterId) (res []domain.AppGroup, err error)
	Get(id domain.AppGroupId) (domain.AppGroup, error)
	Create(dto domain.AppGroup) (id domain.AppGroupId, err error)
	Delete(id domain.AppGroupId) error
	GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (applications []domain.Application, err error)
	UpsertApplication(dto domain.Application) error
	InitWorkflow(appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error
}

type AppGroupRepository struct {
	db *gorm.DB
}

func NewAppGroupRepository(db *gorm.DB) IAppGroupRepository {
	return &AppGroupRepository{
		db: db,
	}
}

// Models
type AppGroup struct {
	gorm.Model

	ID           domain.AppGroupId `gorm:"primarykey"`
	AppGroupType domain.AppGroupType
	ClusterId    domain.ClusterId
	Name         string
	Description  string
	WorkflowId   string
	Status       domain.AppGroupStatus
	StatusDesc   string
	CreatorId    *uuid.UUID `gorm:"type:uuid"`
	Creator      User       `gorm:"foreignKey:CreatorId"`
	UpdatorId    *uuid.UUID `gorm:"type:uuid"`
	Updator      User       `gorm:"foreignKey:UpdatorId"`
}

func (c *AppGroup) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = domain.AppGroupId(helper.GenerateApplicaionGroupId())
	return nil
}

type Application struct {
	gorm.Model

	ID         uuid.UUID `gorm:"primarykey;type:uuid"`
	AppGroupId domain.AppGroupId
	Endpoint   string
	Metadata   datatypes.JSON
	Type       domain.ApplicationType
}

func (c *Application) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *AppGroupRepository) Fetch(clusterId domain.ClusterId) (out []domain.AppGroup, err error) {
	var appGroups []AppGroup
	out = []domain.AppGroup{}

	res := r.db.Find(&appGroups, "cluster_id = ?", clusterId)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, appGroup := range appGroups {
		outAppGroup := reflectAppGroup(appGroup)
		out = append(out, outAppGroup)
	}
	return out, nil
}

func (r *AppGroupRepository) Get(id domain.AppGroupId) (domain.AppGroup, error) {
	var appGroup AppGroup
	res := r.db.First(&appGroup, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.AppGroup{}, fmt.Errorf("Not found appGroup for %s", id)
	}
	resAppGroup := reflectAppGroup(appGroup)
	return resAppGroup, nil
}

func (r *AppGroupRepository) Create(dto domain.AppGroup) (appGroupId domain.AppGroupId, err error) {
	appGroup := AppGroup{
		ClusterId:    dto.ClusterId,
		AppGroupType: dto.AppGroupType,
		Name:         dto.Name,
		Description:  dto.Description,
		Status:       domain.AppGroupStatus_PENDING,
		CreatorId:    &dto.CreatorId,
		UpdatorId:    nil,
	}
	res := r.db.Create(&appGroup)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return appGroup.ID, nil
}

func (r *AppGroupRepository) Delete(id domain.AppGroupId) error {
	res := r.db.Unscoped().Delete(&AppGroup{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("could not delete appGroup %s", id)
	}
	return nil
}

func (r *AppGroupRepository) GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error) {
	var applications []Application
	res := r.db.Where("app_group_id = ? AND type = ?", id, applicationType).Find(&applications)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, application := range applications {
		outApplication := reflectApplication(application)
		out = append(out, outApplication)
	}
	return out, nil
}

func (r *AppGroupRepository) UpsertApplication(dto domain.Application) error {
	res := r.db.Where(Application{
		AppGroupId: dto.AppGroupId,
		Type:       dto.ApplicationType,
	}).
		Assign(Application{
			Endpoint: dto.Endpoint,
			Metadata: datatypes.JSON([]byte(dto.Metadata))}).
		FirstOrCreate(&Application{})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *AppGroupRepository) InitWorkflow(appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error {
	res := r.db.Model(&AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}

	return nil
}

func reflectAppGroup(appGroup AppGroup) domain.AppGroup {
	return domain.AppGroup{
		ID:                appGroup.ID,
		ClusterId:         appGroup.ClusterId,
		AppGroupType:      appGroup.AppGroupType,
		Name:              appGroup.Name,
		Description:       appGroup.Description,
		Status:            appGroup.Status,
		StatusDescription: appGroup.StatusDesc,
		CreatedAt:         appGroup.CreatedAt,
		UpdatedAt:         appGroup.UpdatedAt,
		CreatorId:         *appGroup.CreatorId,
		Creator:           reflectUser(appGroup.Creator),
		UpdatorId:         *appGroup.UpdatorId,
		Updator:           reflectUser(appGroup.Updator),
	}
}

func reflectApplication(application Application) domain.Application {
	return domain.Application{
		ID:              application.ID,
		AppGroupId:      application.AppGroupId,
		ApplicationType: application.Type,
		Endpoint:        application.Endpoint,
		Metadata:        application.Metadata.String(),
		CreatedAt:       application.CreatedAt,
		UpdatedAt:       application.UpdatedAt,
	}
}
