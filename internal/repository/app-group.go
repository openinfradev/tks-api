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
	Fetch(clusterId string) (res []domain.AppGroup, err error)
	Get(id string) (domain.AppGroup, error)
	Create(clusterId string, name string, appGroupType string, creator uuid.UUID, description string) (appGroupId string, err error)
	Delete(id string) error
	GetApplications(appGroupID string, applicationType string) (applications []domain.Application, err error)
	GetApplication(appGroupId string, applicationType string) (out domain.Application, err error)
	UpsertApplication(appGroupID string, appType string, endpoint, metadata string) error
	InitWorkflow(appGroupId string, workflowId string, status domain.AppGroupStatus) error
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

	ID           string `gorm:"primarykey"`
	AppGroupType string
	ClusterId    string
	Name         string
	Creator      uuid.UUID
	Description  string
	WorkflowId   string
	Status       domain.AppGroupStatus
	StatusDesc   string
}

func (c *AppGroup) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateApplicaionGroupId()
	return nil
}

type Application struct {
	gorm.Model

	ID         uuid.UUID `gorm:"primarykey;type:uuid"`
	AppGroupId string
	Endpoint   string
	Metadata   datatypes.JSON
	Type       string
}

func (c *Application) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *AppGroupRepository) Fetch(clusterId string) (out []domain.AppGroup, err error) {
	var appGroups []AppGroup
	out = []domain.AppGroup{}

	res := r.db.Find(&appGroups, "cluster_id = ?", clusterId)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, appGroup := range appGroups {
		outAppGroup := r.reflect(appGroup)
		out = append(out, outAppGroup)
	}
	return out, nil
}

func (r *AppGroupRepository) Get(id string) (domain.AppGroup, error) {
	var appGroup AppGroup
	res := r.db.First(&appGroup, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.AppGroup{}, fmt.Errorf("Not found appGroup for %s", id)
	}
	resAppGroup := r.reflect(appGroup)
	return resAppGroup, nil
}

func (r *AppGroupRepository) Create(clusterId string, name string, appGroupType string, creator uuid.UUID, description string) (appGroupId string, err error) {
	appGroup := AppGroup{
		ClusterId:    clusterId,
		AppGroupType: appGroupType,
		Name:         name,
		Creator:      creator,
		Description:  description,
		Status:       domain.AppGroupStatus_PENDING,
	}
	res := r.db.Create(&appGroup)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return appGroup.ID, nil
}

func (r *AppGroupRepository) Delete(appGroupId string) error {
	res := r.db.Unscoped().Delete(&AppGroup{}, "id = ?", appGroupId)
	if res.Error != nil {
		return fmt.Errorf("could not delete appGroup %s", appGroupId)
	}
	return nil
}

func (r *AppGroupRepository) GetApplications(appGroupId string, applicationType string) (out []domain.Application, err error) {
	var applications []Application
	res := r.db.Where("app_group_id = ? AND type = ?", appGroupId, applicationType).Find(&applications)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, application := range applications {
		outApplication := r.reflectApplication(application)
		out = append(out, outApplication)
	}
	return out, nil
}

func (r *AppGroupRepository) GetApplication(appGroupId string, applicationType string) (out domain.Application, err error) {
	var application Application
	res := r.db.Where("app_group_id = ? AND type = ?", appGroupId, applicationType).First(&application)
	if res.Error != nil {
		return domain.Application{}, res.Error
	}
	return r.reflectApplication(application), nil
}

func (r *AppGroupRepository) UpsertApplication(appGroupId string, appType string, endpoint, metadata string) error {
	res := r.db.Where(Application{
		AppGroupId: appGroupId,
		Type:       appType,
	}).
		Assign(Application{
			Endpoint: endpoint,
			Metadata: datatypes.JSON([]byte(metadata))}).
		FirstOrCreate(&Application{})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *AppGroupRepository) InitWorkflow(appGroupId string, workflowId string, status domain.AppGroupStatus) error {
	res := r.db.Model(&AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}

	return nil
}

func (r *AppGroupRepository) reflect(appGroup AppGroup) domain.AppGroup {

	return domain.AppGroup{
		ID:                appGroup.ID,
		ClusterId:         appGroup.ClusterId,
		AppGroupType:      appGroup.AppGroupType,
		Name:              appGroup.Name,
		Description:       appGroup.Description,
		Status:            appGroup.Status.String(),
		StatusDescription: appGroup.StatusDesc,
		Creator:           appGroup.Creator.String(),
		CreatedAt:         appGroup.CreatedAt,
		UpdatedAt:         appGroup.UpdatedAt,
	}
}

func (r *AppGroupRepository) reflectApplication(application Application) domain.Application {

	return domain.Application{
		ID:              application.ID.String(),
		AppGroupId:      application.AppGroupId,
		ApplicationType: application.Type,
		Endpoint:        application.Endpoint,
		Metadata:        application.Metadata.String(),
		CreatedAt:       application.CreatedAt,
		UpdatedAt:       application.UpdatedAt,
	}
}
