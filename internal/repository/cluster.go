package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IClusterRepository interface {
	WithTrx(*gorm.DB) IClusterRepository
	Fetch(ctx context.Context, pg *pagination.Pagination) (res []model.Cluster, err error)
	FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (res []model.Cluster, err error)
	FetchByOrganizationId(ctx context.Context, organizationId string, userId uuid.UUID, pg *pagination.Pagination) (res []model.Cluster, err error)
	Get(ctx context.Context, id domain.ClusterId) (model.Cluster, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.Cluster, error)
	Create(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error)
	Update(ctx context.Context, dto model.Cluster) (err error)
	Delete(ctx context.Context, id domain.ClusterId) error

	InitWorkflow(ctx context.Context, clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error
	InitWorkflowDescription(ctx context.Context, clusterId domain.ClusterId) error

	SetFavorite(ctx context.Context, clusterId domain.ClusterId, userId uuid.UUID) error
	DeleteFavorite(ctx context.Context, clusterId domain.ClusterId, userId uuid.UUID) error
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

// Logics
func (r *ClusterRepository) WithTrx(trxHandle *gorm.DB) IClusterRepository {
	if trxHandle == nil {
		log.Info(nil, "Transaction Database not found")
		return r
	}
	r.tx = trxHandle
	return r
}

func (r *ClusterRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.Cluster, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.Cluster{}).Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}

	return
}

func (r *ClusterRepository) FetchByOrganizationId(ctx context.Context, organizationId string, userId uuid.UUID, pg *pagination.Pagination) (out []model.Cluster, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.Cluster{}).
		Preload(clause.Associations).
		Joins("left outer join cluster_favorites on clusters.id = cluster_favorites.cluster_id AND cluster_favorites.user_id = ?", userId).
		Where("organization_id = ? AND status != ?", organizationId, domain.ClusterStatus_DELETED).
		Order("cluster_favorites.cluster_id"), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *ClusterRepository) FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []model.Cluster, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.Cluster{}).Preload("CloudAccount").
		Where("cloud_account_id = ?", cloudAccountId), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *ClusterRepository) Get(ctx context.Context, id domain.ClusterId) (out model.Cluster, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", id)
	if res.Error != nil {
		return model.Cluster{}, res.Error
	}
	return
}

func (r *ClusterRepository) GetByName(ctx context.Context, organizationId string, name string) (out model.Cluster, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.Cluster{}, res.Error
	}
	return
}

func (r *ClusterRepository) Create(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error) {
	var cloudAccountId *uuid.UUID
	cloudAccountId = dto.CloudAccountId
	if dto.CloudService == domain.CloudService_BYOH || *dto.CloudAccountId == uuid.Nil {
		cloudAccountId = nil
	}
	cluster := model.Cluster{
		ID:                     domain.ClusterId(helper.GenerateClusterId()),
		OrganizationId:         dto.OrganizationId,
		Name:                   dto.Name,
		Description:            dto.Description,
		CloudAccountId:         cloudAccountId,
		StackTemplateId:        dto.StackTemplateId,
		CreatorId:              dto.CreatorId,
		UpdatorId:              nil,
		Status:                 domain.ClusterStatus_PENDING,
		ClusterType:            dto.ClusterType,
		CloudService:           dto.CloudService,
		ByoClusterEndpointHost: dto.ByoClusterEndpointHost,
		ByoClusterEndpointPort: dto.ByoClusterEndpointPort,
		IsStack:                dto.IsStack,
		TksCpNode:              dto.TksCpNode,
		TksCpNodeMax:           dto.TksCpNodeMax,
		TksCpNodeType:          dto.TksCpNodeType,
		TksInfraNode:           dto.TksInfraNode,
		TksInfraNodeMax:        dto.TksInfraNodeMax,
		TksInfraNodeType:       dto.TksInfraNodeType,
		TksUserNode:            dto.TksUserNode,
		TksUserNodeMax:         dto.TksUserNodeMax,
		TksUserNodeType:        dto.TksUserNodeType,
	}
	if dto.ID != "" {
		cluster.ID = dto.ID
	}

	res := r.db.WithContext(ctx).Create(&cluster)
	if res.Error != nil {
		log.Error(ctx, res.Error)
		return "", res.Error
	}

	return cluster.ID, nil
}

func (r *ClusterRepository) Delete(ctx context.Context, clusterId domain.ClusterId) error {
	res := r.db.WithContext(ctx).Unscoped().Delete(&model.Cluster{}, "id = ?", clusterId)
	if res.Error != nil {
		return fmt.Errorf("could not delete cluster for clusterId %s", clusterId)
	}
	return nil
}

func (r *ClusterRepository) Update(ctx context.Context, dto model.Cluster) error {
	res := r.db.WithContext(ctx).Model(&model.Cluster{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "UpdatorId": dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *ClusterRepository) InitWorkflow(ctx context.Context, clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error {
	res := r.db.WithContext(ctx).Model(&model.Cluster{}).
		Where("ID = ?", clusterId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId, "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster with id %s", clusterId)
	}

	return nil
}

func (r *ClusterRepository) InitWorkflowDescription(ctx context.Context, clusterId domain.ClusterId) error {
	res := r.db.WithContext(ctx).Model(&model.AppGroup{}).
		Where("id = ?", clusterId).
		Updates(map[string]interface{}{"WorkflowId": "", "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster status with id %s", clusterId)
	}

	return nil
}

func (r *ClusterRepository) SetFavorite(ctx context.Context, clusterId domain.ClusterId, userId uuid.UUID) error {
	var clusterFavorites []model.ClusterFavorite
	res := r.db.WithContext(ctx).Where("cluster_id = ? AND user_id = ?", clusterId, userId).Find(&clusterFavorites)
	if res.Error != nil {
		log.Info(ctx, res.Error)
		return res.Error
	}

	if len(clusterFavorites) > 0 {
		return nil
	}

	clusterFavorite := model.ClusterFavorite{
		ClusterId: clusterId,
		UserId:    userId,
	}
	resCreate := r.db.Create(&clusterFavorite)
	if resCreate.Error != nil {
		log.Error(ctx, resCreate.Error)
		return fmt.Errorf("could not create cluster favorite for clusterId %s, userId %s", clusterId, userId)
	}

	return nil
}

func (r *ClusterRepository) DeleteFavorite(ctx context.Context, clusterId domain.ClusterId, userId uuid.UUID) error {
	res := r.db.WithContext(ctx).Unscoped().Delete(&model.ClusterFavorite{}, "cluster_id = ? AND user_id = ?", clusterId, userId)
	if res.Error != nil {
		log.Error(ctx, res.Error)
		return fmt.Errorf("could not delete cluster favorite for clusterId %s, userId %s", clusterId, userId)
	}
	return nil
}
