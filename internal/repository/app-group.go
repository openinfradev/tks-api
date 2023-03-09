package repository

import (
	"fmt"

	"github.com/google/uuid"
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
	InitWorkflow(appGroupId string, workflowId string) error
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
	AppGroupType string `gorm:"uniqueIndex:idx_AppGroupType_ClusterId"`
	ClusterId    string `gorm:"uniqueIndex:idx_AppGroupType_ClusterId"`
	Name         string
	WorkflowId   string
	Status       domain.AppGroupStatus
	StatusDesc   string
	Creator      uuid.UUID
	Description  string
	Workflow     Workflow `gorm:"polymorphic:Ref;polymorphicValue:appgroup"`
}

func (c *AppGroup) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateApplicaionGroupId()
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
	appGroup := AppGroup{ClusterId: clusterId, AppGroupType: appGroupType, Name: name, Creator: creator, Description: description}
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

func (r *AppGroupRepository) InitWorkflow(appGroupId string, workflowId string) error {
	/*
		workflow := Workflow{
			RefID:      appGroupId,
			RefType:    "appgroup",
			WorkflowId: workflowId,
			StatusDesc: "INIT",
		}
	*/
	/*
		res := r.db.Create(&workflow)
		if res.Error != nil {
			return res.Error
		}
	*/

	res := r.db.Where(Workflow{RefID: appGroupId, RefType: "appgroup"}).
		Assign(Workflow{RefID: appGroupId, RefType: "appgroup", WorkflowId: workflowId, StatusDesc: "INIT"}).
		FirstOrCreate(&Workflow{})
	if res.Error != nil {
		return res.Error
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
