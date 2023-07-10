package repository

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IClusterRepository interface {
	WithTrx(*gorm.DB) IClusterRepository
	Fetch() (res []domain.Cluster, err error)
	FetchByCloudAccountId(cloudAccountId uuid.UUID) (res []domain.Cluster, err error)
	FetchByOrganizationId(organizationId string, pg *pagination.Pagination) (res []domain.Cluster, err error)
	Get(id domain.ClusterId) (domain.Cluster, error)
	GetByName(organizationId string, name string) (domain.Cluster, error)
	Create(dto domain.Cluster) (clusterId domain.ClusterId, err error)
	Update(dto domain.Cluster) (err error)
	Delete(id domain.ClusterId) error
	InitWorkflow(clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error
	InitWorkflowDescription(clusterId domain.ClusterId) error
}

type ClusterRepository struct {
	db *gorm.DB
	tx *gorm.DB // used only transaction
}

func NewClusterRepository(db *gorm.DB) IClusterRepository {
	return &ClusterRepository{
		db: db,
		tx: db,
	}
}

// Models
type Cluster struct {
	gorm.Model

	ID                  domain.ClusterId `gorm:"primarykey"`
	Name                string
	OrganizationId      string
	Organization        Organization `gorm:"foreignKey:OrganizationId"`
	Description         string
	WorkflowId          string
	Status              domain.ClusterStatus
	StatusDesc          string
	CloudAccountId      uuid.UUID
	CloudAccount        CloudAccount `gorm:"foreignKey:CloudAccountId"`
	StackTemplateId     uuid.UUID
	StackTemplate       StackTemplate `gorm:"foreignKey:StackTemplateId"`
	CpNodeCnt           int
	CpNodeMachineType   string
	TksNodeCnt          int
	TksNodeMachineType  string
	UserNodeCnt         int
	UserNodeMachineType string
	CreatorId           *uuid.UUID `gorm:"type:uuid"`
	Creator             User       `gorm:"foreignKey:CreatorId"`
	UpdatorId           *uuid.UUID `gorm:"type:uuid"`
	Updator             User       `gorm:"foreignKey:UpdatorId"`
}

func (c *Cluster) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = domain.ClusterId(helper.GenerateClusterId())
	return nil
}

// Logics
func (r *ClusterRepository) WithTrx(trxHandle *gorm.DB) IClusterRepository {
	if trxHandle == nil {
		log.Info("Transaction Database not found")
		return r
	}
	r.tx = trxHandle
	return r
}

func (r *ClusterRepository) Fetch() (out []domain.Cluster, err error) {
	var clusters []Cluster
	res := r.db.Preload(clause.Associations).Find(&clusters)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, cluster := range clusters {
		outCluster := reflectCluster(cluster)
		out = append(out, outCluster)
	}
	return out, nil
}

func (r *ClusterRepository) FetchByOrganizationId(organizationId string, pg *pagination.Pagination) (out []domain.Cluster, err error) {
	var clusters []Cluster
	var total int64

	if pg == nil {
		*pg = pagination.NewDefaultPagination()
	}

	r.db.Find(&clusters, "organization_id = ? AND status != ?", organizationId, domain.ClusterStatus_DELETED).Count(&total)

	pg.TotalRows = total
	pg.TotalPages = int(math.Ceil(float64(total) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)

	res := r.db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).
		Preload(clause.Associations).Order("updated_at desc, created_at desc").
		Find(&clusters, "organization_id = ? AND status != ?", organizationId, domain.ClusterStatus_DELETED)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return out, nil
	}

	for _, cluster := range clusters {
		outCluster := reflectCluster(cluster)
		out = append(out, outCluster)
	}

	return
}

func (r *ClusterRepository) FetchByCloudAccountId(cloudAccountId uuid.UUID) (out []domain.Cluster, err error) {
	var clusters []Cluster

	res := r.db.Preload("CloudAccount").Order("updated_at desc, created_at desc").Find(&clusters, "cloud_account_id = ?", cloudAccountId)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return out, nil
	}

	for _, cluster := range clusters {
		outCluster := reflectCluster(cluster)
		out = append(out, outCluster)
	}

	return
}

func (r *ClusterRepository) Get(id domain.ClusterId) (out domain.Cluster, err error) {
	var cluster Cluster
	res := r.db.Preload(clause.Associations).First(&cluster, "id = ?", id)
	if res.Error != nil {
		return domain.Cluster{}, res.Error
	}
	out = reflectCluster(cluster)
	return
}

func (r *ClusterRepository) GetByName(organizationId string, name string) (out domain.Cluster, err error) {
	var cluster Cluster
	res := r.db.Preload(clause.Associations).First(&cluster, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return domain.Cluster{}, res.Error
	}
	out = reflectCluster(cluster)
	return
}

func (r *ClusterRepository) Create(dto domain.Cluster) (clusterId domain.ClusterId, err error) {
	cluster := Cluster{
		OrganizationId:      dto.OrganizationId,
		Name:                dto.Name,
		Description:         dto.Description,
		CloudAccountId:      dto.CloudAccountId,
		StackTemplateId:     dto.StackTemplateId,
		CreatorId:           dto.CreatorId,
		UpdatorId:           nil,
		Status:              domain.ClusterStatus_PENDING,
		CpNodeCnt:           dto.Conf.CpNodeCnt,
		CpNodeMachineType:   dto.Conf.CpNodeMachineType,
		TksNodeCnt:          dto.Conf.TksNodeCnt,
		TksNodeMachineType:  dto.Conf.TksNodeMachineType,
		UserNodeCnt:         dto.Conf.UserNodeCnt,
		UserNodeMachineType: dto.Conf.UserNodeMachineType,
	}
	res := r.db.Create(&cluster)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return cluster.ID, nil
}

func (r *ClusterRepository) Delete(clusterId domain.ClusterId) error {
	res := r.db.Unscoped().Delete(&Cluster{}, "id = ?", clusterId)
	if res.Error != nil {
		return fmt.Errorf("could not delete cluster for clusterId %s", clusterId)
	}
	return nil
}

func (r *ClusterRepository) Update(dto domain.Cluster) error {
	res := r.db.Model(&Cluster{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "UpdatorId": dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *ClusterRepository) InitWorkflow(clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error {
	res := r.db.Model(&Cluster{}).
		Where("ID = ?", clusterId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId, "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster with id %s", clusterId)
	}

	return nil
}

func (r *ClusterRepository) InitWorkflowDescription(clusterId domain.ClusterId) error {
	res := r.db.Model(&AppGroup{}).
		Where("id = ?", clusterId).
		Updates(map[string]interface{}{"WorkflowId": "", "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster status with id %s", clusterId)
	}

	return nil
}

func reflectCluster(cluster Cluster) domain.Cluster {
	return domain.Cluster{
		ID:              cluster.ID,
		OrganizationId:  cluster.OrganizationId,
		Name:            cluster.Name,
		Description:     cluster.Description,
		CloudAccountId:  cluster.CloudAccountId,
		CloudAccount:    reflectCloudAccount(cluster.CloudAccount),
		StackTemplateId: cluster.StackTemplateId,
		StackTemplate:   reflectStackTemplate(cluster.StackTemplate),
		Status:          cluster.Status,
		StatusDesc:      cluster.StatusDesc,
		CreatorId:       cluster.CreatorId,
		Creator:         reflectSimpleUser(cluster.Creator),
		UpdatorId:       cluster.UpdatorId,
		Updator:         reflectSimpleUser(cluster.Updator),
		CreatedAt:       cluster.CreatedAt,
		UpdatedAt:       cluster.UpdatedAt,
		Conf: domain.ClusterConf{
			CpNodeCnt:           int(cluster.CpNodeCnt),
			CpNodeMachineType:   cluster.CpNodeMachineType,
			TksNodeCnt:          int(cluster.TksNodeCnt),
			TksNodeMachineType:  cluster.TksNodeMachineType,
			UserNodeCnt:         int(cluster.UserNodeCnt),
			UserNodeMachineType: cluster.UserNodeMachineType,
		},
	}
}

func reflectSimpleCluster(cluster Cluster) domain.Cluster {
	return domain.Cluster{
		ID:             cluster.ID,
		OrganizationId: cluster.OrganizationId,
		Name:           cluster.Name,
	}
}
