package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type ICloudSettingRepository interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	GetByOrganizationId(organizationId string) (domain.CloudSetting, error)
	Create(organizationId string, input domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error)
	Delete(cloudSettingId uuid.UUID) (err error)
}

type CloudSettingRepository struct {
	db *gorm.DB
}

func NewCloudSettingRepository(db *gorm.DB) ICloudSettingRepository {
	return &CloudSettingRepository{
		db: db,
	}
}

// Models
type CloudSetting struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Name           string
	Description    string
	Resource       string
	Type           domain.CloudType
	Creator        uuid.UUID
}

func (c *CloudSetting) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *CloudSettingRepository) Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error) {
	var cloudSetting CloudSetting
	res := r.db.First(&cloudSetting, "id = ?", cloudSettingId)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.CloudSetting{}, fmt.Errorf("Not found cloudSetting for %s", cloudSettingId)
	}
	resCloudSetting := r.reflect(cloudSetting)
	return resCloudSetting, nil
}

func (r *CloudSettingRepository) GetByOrganizationId(organizationId string) (domain.CloudSetting, error) {
	var cloudSetting CloudSetting
	res := r.db.First(&cloudSetting, "organization_id = ?", organizationId)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.CloudSetting{}, fmt.Errorf("Not found cloudSetting for organizationId %s", organizationId)
	}
	resCloudSetting := r.reflect(cloudSetting)
	return resCloudSetting, nil
}

func (r *CloudSettingRepository) Create(organizationId string, input domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error) {
	cloudSetting := CloudSetting{
		OrganizationId: organizationId,
		Name:           input.Name,
		Description:    input.Description,
		Resource:       resource,
		Creator:        creator}
	res := r.db.Create(&cloudSetting)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudSetting.ID, nil
}

func (r *CloudSettingRepository) Delete(cloudSettingId uuid.UUID) (err error) {
	res := r.db.Delete(&CloudSetting{}, "id = ?", cloudSettingId)
	if res.Error != nil {
		return fmt.Errorf("could not delete cloudSetting for cloudSettingId %s", cloudSettingId)
	}
	return nil
}

func (r *CloudSettingRepository) reflect(cloudSetting CloudSetting) domain.CloudSetting {
	return domain.CloudSetting{
		ID:             cloudSetting.ID.String(),
		OrganizationId: cloudSetting.OrganizationId,
		Name:           cloudSetting.Name,
		Description:    cloudSetting.Description,
		Resource:       cloudSetting.Resource,
		Creator:        cloudSetting.Creator.String(),
		CreatedAt:      cloudSetting.CreatedAt,
		UpdatedAt:      cloudSetting.UpdatedAt,
	}
}
