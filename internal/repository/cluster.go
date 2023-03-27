package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
	Get(id string) (domain.Cluster, error)
	Create(organizationId string, templateId string, name string, conf *domain.ClusterConf, creator uuid.UUID, description string) (clusterId string, err error)
	Delete(id string) error
	InitWorkflow(clusterId string, workflowId string) error
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

	ID             string `gorm:"primarykey"`
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
	Creator        uuid.UUID
	Description    string
	WorkflowId     string
	Status         domain.ClusterStatus
	StatusDesc     string
	CloudSettingId uuid.UUID
	CloudSetting   CloudSetting `gorm:"foreignKey:CloudSettingId"`
}

func (c *Cluster) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateClusterId()
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
	out = []domain.Cluster{}

	res := r.db.Find(&clusters)
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

	res := r.db.Find(&clusters, "organization_id = ?", organizationId)

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

	res := r.db.Find(&clusters, "cloud_setting_id = ?", cloudSettingId)

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

func (r *ClusterRepository) Get(id string) (domain.Cluster, error) {
	var cluster Cluster
	res := r.db.First(&cluster, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		log.Info(res.Error)
		return domain.Cluster{}, fmt.Errorf("Not found cluster for %s", id)
	}
	resCluster := r.reflect(cluster)
	return resCluster, nil
}

func (r *ClusterRepository) Create(organizationId string, templateId string, name string, conf *domain.ClusterConf, creator uuid.UUID, description string) (string, error) {
	cluster := Cluster{
		OrganizationId: organizationId,
		TemplateId:     templateId,
		Name:           name,
		Creator:        creator,
		Description:    description,
		SshKeyName:     conf.SshKeyName,
		Region:         conf.Region,
		NumOfAz:        conf.NumOfAz,
		MachineType:    conf.MachineType,
		MinSizePerAz:   conf.MinSizePerAz,
		MaxSizePerAz:   conf.MaxSizePerAz,
	}
	res := r.db.Create(&cluster)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return cluster.ID, nil
}

func (r *ClusterRepository) Delete(clusterId string) error {
	res := r.db.Unscoped().Delete(&Cluster{}, "id = ?", clusterId)
	if res.Error != nil {
		return fmt.Errorf("could not delete cluster for clusterId %s", clusterId)
	}
	return nil
}

func (r *ClusterRepository) InitWorkflow(clusterId string, workflowId string) error {
	res := r.db.Model(&Cluster{}).
		Where("ID = ?", clusterId).
		Updates(map[string]interface{}{"Status": domain.ClusterStatus_INSTALLING, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster with id %s", clusterId)
	}

	return nil

}

func (r *ClusterRepository) reflect(cluster Cluster) domain.Cluster {

	log.Info(helper.ModelToJson(cluster))

	return domain.Cluster{
		ID:             cluster.ID,
		OrganizationId: cluster.OrganizationId,
		Name:           cluster.Name,
		Description:    cluster.Description,
		Status:         cluster.Status.String(),
		StatusDesc:     cluster.StatusDesc,
		Creator:        cluster.Creator.String(),
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
