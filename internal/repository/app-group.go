package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-common/pkg/helper"
	"github.com/openinfradev/tks-common/pkg/log"
)

// Interfaces
type IAppGroupRepository interface {
	Fetch(clusterId string) (res []domain.AppGroup, err error)
	Get(id string) (domain.AppGroup, error)
	Create(clusterId string, name string, appGroupType string, creator uuid.UUID, description string) (appGroupId string, err error)
	Delete(id string) error
	UpdateAppGroupStatus(appGroupId string, status domain.AppGroupStatus, workflowId string) error
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
	Name         string
	ClusterId    string
	WorkflowId   string
	Status       domain.AppGroupStatus
	StatusDesc   string
	Creator      uuid.UUID
	Description  string
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

func (c *AppGroup) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateApplicaionGroupId()
	return nil
}

// Logics
func (r *AppGroupRepository) Fetch(clusterId string) (out []domain.AppGroup, err error) {
	var appGroups []AppGroup
	out = []domain.AppGroup{}

	res := r.db.Find(&appGroups)
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
	res := r.db.Delete(&AppGroup{}, "id = ?", appGroupId)
	if res.Error != nil {
		return fmt.Errorf("could not delete appGroup %s", appGroupId)
	}
	return nil
}

func (r *AppGroupRepository) UpdateAppGroupStatus(appGroupId string, status domain.AppGroupStatus, workflowId string) error {
	res := r.db.Model(&AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appGroup with id %s", appGroupId)
	}

	return nil
}

func (r *AppGroupRepository) reflect(appGroup AppGroup) domain.AppGroup {

	return domain.AppGroup{
		Id:                appGroup.ID,
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
