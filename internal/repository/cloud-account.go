package repository

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type ICloudAccountRepository interface {
	Get(cloudAccountId uuid.UUID) (domain.CloudAccount, error)
	GetByName(organizationId string, name string) (domain.CloudAccount, error)
	GetByAwsAccountId(awsAccountId string) (domain.CloudAccount, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]domain.CloudAccount, error)
	Create(dto domain.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(dto domain.CloudAccount) (err error)
	Delete(cloudAccountId uuid.UUID) (err error)
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
	Name           string       `gorm:"index"`
	Description    string       `gorm:"index"`
	Resource       string
	CloudService   string
	WorkflowId     string
	Status         domain.CloudAccountStatus
	StatusDesc     string
	AwsAccountId   string
	CreatedIAM     bool
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

func (r *CloudAccountRepository) GetByAwsAccountId(awsAccountId string) (out domain.CloudAccount, err error) {
	var cloudAccount CloudAccount
	res := r.db.Preload(clause.Associations).First(&cloudAccount, "aws_account_id = ? AND status != ?", awsAccountId, domain.CloudAccountStatus_DELETED)

	if res.Error != nil {
		return domain.CloudAccount{}, res.Error
	}
	out = reflectCloudAccount(cloudAccount)
	return
}

func (r *CloudAccountRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []domain.CloudAccount, err error) {
	var cloudAccounts []CloudAccount
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("cloud_accounts", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&CloudAccount{}).
		Preload(clause.Associations).
		Where("organization_id = ? AND status != ?", organizationId, domain.CloudAccountStatus_DELETED))
	db.Count(&pg.TotalRows)

	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))
	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&cloudAccounts)
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
		AwsAccountId:   dto.AwsAccountId,
		CreatedIAM:     false,
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

func (r *CloudAccountRepository) Delete(cloudAccountId uuid.UUID) (err error) {
	res := r.db.Delete(&CloudAccount{}, "id = ?", cloudAccountId)
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

func reflectCloudAccount(cloudAccount CloudAccount) (out domain.CloudAccount) {
	if err := serializer.Map(cloudAccount, &out); err != nil {
		log.Error(err)
	}
	return
}
