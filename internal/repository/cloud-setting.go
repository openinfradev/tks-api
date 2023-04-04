package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type ICloudSettingRepository interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	Fetch(organizationId string) ([]domain.CloudSetting, error)
	Create(dto domain.CloudSetting) (cloudSettingId uuid.UUID, err error)
	Update(dto domain.CloudSetting) (err error)
	Delete(dto domain.CloudSetting) (err error)
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

	for _, cloudSetting := range cloudSettings {
		out = append(out, r.reflect(cloudSetting))
	}
	return
}

func (r *CloudSettingRepository) Create(dto domain.CloudSetting) (cloudSettingId uuid.UUID, err error) {
	cloudSetting := CloudSetting{
		OrganizationId: dto.OrganizationId,
		Name:           dto.Name,
		Description:    dto.Description,
		Type:           dto.Type,
		Resource:       dto.Resource,
		Creator:        dto.Creator}
	res := r.db.Create(&cloudSetting)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudSetting.ID, nil
}

func (r *CloudSettingRepository) Update(dto domain.CloudSetting) (err error) {
	res := r.db.Model(&CloudSetting{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "Resource": dto.Resource, "Updator": dto.Updator})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudSettingRepository) Delete(dto domain.CloudSetting) (err error) {
	res := r.db.Delete(&CloudSetting{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudSettingRepository) reflect(cloudSetting CloudSetting) domain.CloudSetting {
	return domain.CloudSetting{
		ID:             cloudSetting.ID,
		OrganizationId: cloudSetting.OrganizationId,
		Name:           cloudSetting.Name,
		Description:    cloudSetting.Description,
		Resource:       cloudSetting.Resource,
		Type:           cloudSetting.Type,
		Creator:        cloudSetting.Creator,
		Updator:        cloudSetting.Updator,
		CreatedAt:      cloudSetting.CreatedAt,
		UpdatedAt:      cloudSetting.UpdatedAt,
	}
}
