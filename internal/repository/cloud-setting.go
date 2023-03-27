package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

// Interfaces
type ICloudSettingRepository interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	Fetch(organizationId string) ([]domain.CloudSetting, error)
	Create(organizationId string, input domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error)
	Update(cloudSettingId uuid.UUID, input domain.UpdateCloudSettingRequest, resource string, updator uuid.UUID) (err error)
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
	Updator        uuid.UUID
}

func (c *CloudSetting) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *CloudSettingRepository) Get(cloudSettingId uuid.UUID) (out domain.CloudSetting, err error) {
	var cloudSetting CloudSetting
	res := r.db.First(&cloudSetting, "id = ?", cloudSettingId)
	if res.Error != nil {
		return domain.CloudSetting{}, res.Error
	}
	out = r.reflect(cloudSetting)
	return
}

func (r *CloudSettingRepository) Fetch(organizationId string) (out []domain.CloudSetting, err error) {
	var cloudSettings []CloudSetting
	res := r.db.Find(&cloudSettings, "organization_id = ?", organizationId)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("No data found"))
	}

	for _, cloudSetting := range cloudSettings {
		out = append(out, r.reflect(cloudSetting))
	}
	return
}

func (r *CloudSettingRepository) Create(organizationId string, input domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error) {
	cloudSetting := CloudSetting{
		OrganizationId: organizationId,
		Name:           input.Name,
		Description:    input.Description,
		Type:           input.Type,
		Resource:       resource,
		Creator:        creator}
	res := r.db.Create(&cloudSetting)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudSetting.ID, nil
}

func (r *CloudSettingRepository) Update(cloudSettingId uuid.UUID, in domain.UpdateCloudSettingRequest, resource string, updator uuid.UUID) (err error) {
	res := r.db.Model(&CloudSetting{}).
		Where("id = ?", cloudSettingId).
		Updates(map[string]interface{}{"Description": in.Description, "Resource": resource, "Updator": updator})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudSettingRepository) Delete(cloudSettingId uuid.UUID) (err error) {
	res := r.db.Delete(&CloudSetting{}, "id = ?", cloudSettingId)
	if res.Error != nil {
		return res.Error
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
		Type:           cloudSetting.Type,
		Creator:        cloudSetting.Creator.String(),
		Updator:        cloudSetting.Updator.String(),
		CreatedAt:      cloudSetting.CreatedAt,
		UpdatedAt:      cloudSetting.UpdatedAt,
	}
}
