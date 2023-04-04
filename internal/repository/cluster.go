package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IClusterRepository interface {
	WithTrx(*gorm.DB) IClusterRepository
	Fetch() (res []domain.Cluster, err error)
	FetchByOrganizationId(organizationId string) (res []domain.Cluster, err error)
	FetchByCloudSettingId(cloudSettingId uuid.UUID) (res []domain.Cluster, err error)
	Get(id domain.ClusterId) (domain.Cluster, error)
	Create(dto domain.Cluster) (clusterId domain.ClusterId, err error)
	Delete(id domain.ClusterId) error
	InitWorkflow(clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error
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

	ID             domain.ClusterId `gorm:"primarykey"`
	Name           string
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	TemplateId     string
	SshKeyName     string
	Region         string
	NumOfAz        int
	MachineType    string
	MinSizePerAz   int
	MaxSizePerAz   int
	Description    string
	WorkflowId     string
	Status         domain.ClusterStatus
	StatusDesc     string
	CloudSettingId uuid.UUID
	CloudSetting   CloudSetting `gorm:"foreignKey:CloudSettingId"`
	CreatorId      *uuid.UUID   `gorm:"type:uuid"`
	Creator        User         `gorm:"foreignKey:CreatorId"`
	UpdatorId      *uuid.UUID   `gorm:"type:uuid"`
	Updator        User         `gorm:"foreignKey:UpdatorId"`
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
		outCluster := r.reflect(cluster)
		out = append(out, outCluster)
	}
	return out, nil
}

// [TODO] Need refactoring about filters and pagination
func (r *ClusterRepository) FetchByOrganizationId(organizationId string) (out []domain.Cluster, err error) {
	var clusters []Cluster

	res := r.db.Preload("CloudSetting").Find(&clusters, "organization_id = ?", organizationId)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return out, nil
	}

	for _, cluster := range clusters {
		outCluster := r.reflect(cluster)
		out = append(out, outCluster)
	}

	return
}

func (r *ClusterRepository) FetchByCloudSettingId(cloudSettingId uuid.UUID) (out []domain.Cluster, err error) {
	var clusters []Cluster

	res := r.db.Preload("CloudSetting").Find(&clusters, "cloud_setting_id = ?", cloudSettingId)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return out, nil
	}

	for _, cluster := range clusters {
		outCluster := r.reflect(cluster)
		out = append(out, outCluster)
	}

	return
}

func (r *ClusterRepository) Get(id domain.ClusterId) (domain.Cluster, error) {
	var cluster Cluster
	res := r.db.Preload(clause.Associations).First(&cluster, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		log.Info(res.Error)
		return domain.Cluster{}, fmt.Errorf("Not found cluster for %s", id)
	}
	resCluster := r.reflect(cluster)
	return resCluster, nil
}

func (r *ClusterRepository) Create(dto domain.Cluster) (clusterId domain.ClusterId, err error) {
	cluster := Cluster{
		OrganizationId: dto.OrganizationId,
		TemplateId:     dto.TemplateId,
		Name:           dto.Name,
		Description:    dto.Description,
		CloudSettingId: dto.CloudSettingId,
		CreatorId:      dto.CreatorId,
		UpdatorId:      nil,
		SshKeyName:     dto.Conf.SshKeyName,
		Region:         dto.Conf.Region,
		NumOfAz:        dto.Conf.NumOfAz,
		MachineType:    dto.Conf.MachineType,
		MinSizePerAz:   dto.Conf.MinSizePerAz,
		MaxSizePerAz:   dto.Conf.MaxSizePerAz,
		Status:         domain.ClusterStatus_PENDING,
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

func (r *ClusterRepository) InitWorkflow(clusterId domain.ClusterId, workflowId string, status domain.ClusterStatus) error {
	res := r.db.Model(&Cluster{}).
		Where("ID = ?", clusterId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster with id %s", clusterId)
	}

	return nil

}

func (r *ClusterRepository) reflect(cluster Cluster) domain.Cluster {
	return domain.Cluster{
		ID:             cluster.ID,
		OrganizationId: cluster.OrganizationId,
		Name:           cluster.Name,
		Description:    cluster.Description,
		CloudSetting:   reflectCloudSetting(cluster.CloudSetting),
		Status:         cluster.Status,
		StatusDesc:     cluster.StatusDesc,
		CreatorId:      cluster.CreatorId,
		Creator:        reflectUser(cluster.Creator),
		UpdatorId:      cluster.UpdatorId,
		Updator:        reflectUser(cluster.Updator),
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
		Conf: domain.ClusterConf{
			SshKeyName:   cluster.SshKeyName,
			Region:       cluster.Region,
			MachineType:  cluster.MachineType,
			NumOfAz:      int(cluster.NumOfAz),
			MinSizePerAz: int(cluster.MinSizePerAz),
			MaxSizePerAz: int(cluster.MaxSizePerAz),
		},
	}
}
