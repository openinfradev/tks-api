package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type ICloudAccountRepository interface {
	Get(cloudAccountId uuid.UUID) (domain.CloudAccount, error)
	GetByName(organizationId string, name string) (domain.CloudAccount, error)
	Fetch(organizationId string) ([]domain.CloudAccount, error)
	Create(dto domain.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(dto domain.CloudAccount) (err error)
	Delete(dto domain.CloudAccount) (err error)
	InitWorkflow(cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) (err error)
}

type CloudAccountRepository struct {
	db *gorm.DB
}

func NewCloudAccountRepository(db *gorm.DB) ICloudAccountRepository {
	return &CloudAccountRepository{
		db: db,
	}
}

// Models
type CloudAccount struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Name           string
	Description    string
	Resource       string
	CloudService   string
	WorkflowId     string
	Status         domain.CloudAccountStatus
	StatusDesc     string
	CreatorId      *uuid.UUID `gorm:"type:uuid"`
	Creator        User       `gorm:"foreignKey:CreatorId"`
	UpdatorId      *uuid.UUID `gorm:"type:uuid"`
	Updator        User       `gorm:"foreignKey:UpdatorId"`
}

func (c *CloudAccount) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *CloudAccountRepository) Get(cloudAccountId uuid.UUID) (out domain.CloudAccount, err error) {
	var cloudAccount CloudAccount
	res := r.db.Preload(clause.Associations).First(&cloudAccount, "id = ?", cloudAccountId)
	if res.Error != nil {
		return domain.CloudAccount{}, res.Error
	}
	out = reflectCloudAccount(cloudAccount)
	return
}

func (r *CloudAccountRepository) GetByName(organizationId string, name string) (out domain.CloudAccount, err error) {
	var cloudAccount CloudAccount
	res := r.db.Preload(clause.Associations).First(&cloudAccount, "organization_id = ? AND name = ?", organizationId, name)

	if res.Error != nil {
		return domain.CloudAccount{}, res.Error
	}
	out = reflectCloudAccount(cloudAccount)
	return
}

func (r *CloudAccountRepository) Fetch(organizationId string) (out []domain.CloudAccount, err error) {
	var cloudAccounts []CloudAccount
	res := r.db.Preload(clause.Associations).Find(&cloudAccounts, "organization_id = ?", organizationId)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, cloudAccount := range cloudAccounts {
		out = append(out, reflectCloudAccount(cloudAccount))
	}
	return
}

func (r *CloudAccountRepository) Create(dto domain.CloudAccount) (cloudAccountId uuid.UUID, err error) {
	cloudAccount := CloudAccount{
		OrganizationId: dto.OrganizationId,
		Name:           dto.Name,
		Description:    dto.Description,
		CloudService:   dto.CloudService,
		Resource:       dto.Resource,
		Status:         domain.CloudAccountStatus_PENDING,
		CreatorId:      &dto.CreatorId}
	res := r.db.Create(&cloudAccount)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudAccount.ID, nil
}

func (r *CloudAccountRepository) Update(dto domain.CloudAccount) (err error) {
	res := r.db.Model(&CloudAccount{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "Resource": dto.Resource, "UpdatorId": dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) Delete(dto domain.CloudAccount) (err error) {
	res := r.db.Delete(&CloudAccount{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) InitWorkflow(cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) error {
	res := r.db.Model(&CloudAccount{}).
		Where("ID = ?", cloudAccountId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cloud-account with id %s", &cloudAccountId)
	}

	return nil
}

func reflectCloudAccount(cloudAccount CloudAccount) domain.CloudAccount {
	return domain.CloudAccount{
		ID:             cloudAccount.ID,
		OrganizationId: cloudAccount.OrganizationId,
		Name:           cloudAccount.Name,
		Description:    cloudAccount.Description,
		Resource:       cloudAccount.Resource,
		CloudService:   cloudAccount.CloudService,
		Status:         cloudAccount.Status,
		StatusDesc:     cloudAccount.StatusDesc,
		Creator:        reflectSimpleUser(cloudAccount.Creator),
		Updator:        reflectSimpleUser(cloudAccount.Updator),
		CreatedAt:      cloudAccount.CreatedAt,
		UpdatedAt:      cloudAccount.UpdatedAt,
	}
}
