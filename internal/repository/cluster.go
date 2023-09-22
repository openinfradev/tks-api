package repository

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IClusterRepository interface {
	WithTrx(*gorm.DB) IClusterRepository
	Fetch(pg *pagination.Pagination) (res []domain.Cluster, err error)
	FetchByCloudAccountId(cloudAccountId uuid.UUID, pg *pagination.Pagination) (res []domain.Cluster, err error)
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

	ID               domain.ClusterId `gorm:"primarykey"`
	Name             string           `gorm:"index"`
	OrganizationId   string
	Organization     Organization `gorm:"foreignKey:OrganizationId"`
	Description      string       `gorm:"index"`
	WorkflowId       string
	Status           domain.ClusterStatus
	StatusDesc       string
	CloudAccountId   uuid.UUID
	CloudAccount     CloudAccount `gorm:"foreignKey:CloudAccountId"`
	StackTemplateId  uuid.UUID
	StackTemplate    StackTemplate `gorm:"foreignKey:StackTemplateId"`
	TksCpNode        int
	TksCpNodeMax     int
	TksCpNodeType    string
	TksInfraNode     int
	TksInfraNodeMax  int
	TksInfraNodeType string
	TksUserNode      int
	TksUserNodeMax   int
	TksUserNodeType  string
	CreatorId        *uuid.UUID `gorm:"type:uuid"`
	Creator          User       `gorm:"foreignKey:CreatorId"`
	UpdatorId        *uuid.UUID `gorm:"type:uuid"`
	Updator          User       `gorm:"foreignKey:UpdatorId"`
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

func (r *ClusterRepository) Fetch(pg *pagination.Pagination) (out []domain.Cluster, err error) {
	var clusters []Cluster
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("clusters", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&Cluster{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&clusters)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, cluster := range clusters {
		outCluster := reflectCluster(cluster)
		out = append(out, outCluster)
	}
	return
}

func (r *ClusterRepository) FetchByOrganizationId(organizationId string, pg *pagination.Pagination) (out []domain.Cluster, err error) {
	var clusters []Cluster
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	pg.SortColumn = "created_at"
	pg.SortOrder = "DESC"
	filterFunc := CombinedGormFilter("clusters", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&Cluster{}).Preload(clause.Associations).
		Where("organization_id = ? AND status != ?", organizationId, domain.ClusterStatus_DELETED))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&clusters)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, cluster := range clusters {
		outCluster := reflectCluster(cluster)
		out = append(out, outCluster)
	}
	return
}

func (r *ClusterRepository) FetchByCloudAccountId(cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []domain.Cluster, err error) {
	var clusters []Cluster
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	pg.SortColumn = "created_at"
	pg.SortOrder = "DESC"
	filterFunc := CombinedGormFilter("clusters", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&Cluster{}).Preload("CloudAccount").
		Where("cloud_account_id = ?", cloudAccountId))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&clusters)
	if res.Error != nil {
		return nil, res.Error
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
		OrganizationId:   dto.OrganizationId,
		Name:             dto.Name,
		Description:      dto.Description,
		CloudAccountId:   dto.CloudAccountId,
		StackTemplateId:  dto.StackTemplateId,
		CreatorId:        dto.CreatorId,
		UpdatorId:        nil,
		Status:           domain.ClusterStatus_PENDING,
		TksCpNode:        dto.Conf.TksCpNode,
		TksCpNodeMax:     dto.Conf.TksCpNodeMax,
		TksCpNodeType:    dto.Conf.TksCpNodeType,
		TksInfraNode:     dto.Conf.TksInfraNode,
		TksInfraNodeMax:  dto.Conf.TksInfraNodeMax,
		TksInfraNodeType: dto.Conf.TksInfraNodeType,
		TksUserNode:      dto.Conf.TksUserNode,
		TksUserNodeMax:   dto.Conf.TksUserNodeMax,
		TksUserNodeType:  dto.Conf.TksUserNodeType,
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

func reflectCluster(cluster Cluster) (out domain.Cluster) {
	if err := serializer.Map(cluster.Model, &out); err != nil {
		log.Error(err)
	}
	if err := serializer.Map(cluster, &out); err != nil {
		log.Error(err)
	}

	if err := serializer.Map(cluster, &out.Conf); err != nil {
		log.Error(err)
	}
	return
}
