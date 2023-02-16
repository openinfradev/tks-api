package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-common/pkg/helper"
	"github.com/openinfradev/tks-common/pkg/log"
)

// Interfaces
type IClusterRepository interface {
	WithTrx(*gorm.DB) IClusterRepository
	Fetch() (res []domain.Cluster, err error)
	FetchByContractId(contractId string) (res []domain.Cluster, err error)
	Get(id string) (domain.Cluster, error)
	Create(contractId string, templateId string, name string, conf *domain.ClusterConf, creator uuid.UUID, description string) (clusterId string, err error)
	Delete(id string) error
	UpdateClusterStatus(clusterId string, status domain.ClusterStatus, workflowId string) error
}

type ClusterRepository struct {
	db *gorm.DB
	tx *gorm.DB // used only transaction
}

func NewClusterRepository(db *gorm.DB) IClusterRepository {
	return &ClusterRepository{
		db: db,
	}
}

// Models
type Cluster struct {
	gorm.Model
	ID           string `gorm:"primarykey"`
	Name         string
	ContractId   string
	TemplateId   string
	WorkflowId   string
	Status       domain.ClusterStatus
	StatusDesc   string
	SshKeyName   string
	Region       string
	NumOfAz      int32
	MachineType  string
	MinSizePerAz int32
	MaxSizePerAz int32
	Creator      uuid.UUID
	Description  string
	UpdatedAt    time.Time
	CreatedAt    time.Time
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

func (r *ClusterRepository) FetchByContractId(contractId string) (resClusters []domain.Cluster, err error) {
	var clusters []Cluster

	res := r.db.Find(&clusters, "contract_id = ?", contractId)

	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding clusters with contractID: %s", contractId)
	}

	if res.RowsAffected == 0 {
		return resClusters, nil
	}

	for _, cluster := range clusters {
		outCluster := r.reflect(cluster)
		resClusters = append(resClusters, outCluster)
	}

	return resClusters, nil
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

func (r *ClusterRepository) Create(contractId string, templateId string, name string, conf *domain.ClusterConf, creator uuid.UUID, description string) (string, error) {
	cluster := Cluster{ContractId: contractId, TemplateId: templateId, Name: name, Creator: creator, Description: description}
	res := r.db.Create(&cluster)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return cluster.ID, nil
}

func (r *ClusterRepository) Delete(clusterId string) error {
	res := r.db.Delete(&Cluster{}, "id = ?", clusterId)
	if res.Error != nil {
		return fmt.Errorf("could not delete cluster for clusterId %s", clusterId)
	}
	return nil
}

func (r *ClusterRepository) UpdateClusterStatus(clusterId string, status domain.ClusterStatus, workflowId string) error {
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
		Id:                cluster.ID,
		ProjectId:         cluster.ContractId,
		Name:              cluster.Name,
		Description:       cluster.Description,
		Status:            cluster.Status.String(),
		StatusDescription: cluster.StatusDesc,
		Creator:           cluster.Creator.String(),
		CreatedAt:         cluster.CreatedAt,
		UpdatedAt:         cluster.UpdatedAt,
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
