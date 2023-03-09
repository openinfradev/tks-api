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
	Status         domain.ClusterStatus
	SshKeyName     string
	Region         string
	NumOfAz        int32
	MachineType    string
	MinSizePerAz   int32
	MaxSizePerAz   int32
	Creator        uuid.UUID
	Description    string
	Workflow       Workflow `gorm:"polymorphic:Ref;polymorphicValue:cluster"`
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

func (r *ClusterRepository) FetchByOrganizationId(organizationId string) (resClusters []domain.Cluster, err error) {
	var clusters []Cluster

	res := r.db.Preload("Workflow").Find(&clusters, "organization_id = ?", organizationId)

	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding clusters with organizationId: %s", organizationId)
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
	res := r.db.Preload("Workflow").First(&cluster, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		log.Info(res.Error)
		return domain.Cluster{}, fmt.Errorf("Not found cluster for %s", id)
	}
	resCluster := r.reflect(cluster)
	return resCluster, nil
}

func (r *ClusterRepository) Create(organizationId string, templateId string, name string, conf *domain.ClusterConf, creator uuid.UUID, description string) (string, error) {
	cluster := Cluster{OrganizationId: organizationId, TemplateId: templateId, Name: name, Creator: creator, Description: description}
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
	workflow := Workflow{
		RefID:      clusterId,
		RefType:    "cluster",
		WorkflowId: workflowId,
		StatusDesc: "INIT",
	}
	res := r.db.Create(&workflow)
	if res.Error != nil {
		return res.Error
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
