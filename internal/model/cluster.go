package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

// Models
type Cluster struct {
	gorm.Model

	ID                     domain.ClusterId `gorm:"primarykey"`
	Name                   string           `gorm:"index"`
	CloudService           string           `gorm:"default:AWS"`
	OrganizationId         string
	Organization           Organization `gorm:"foreignKey:OrganizationId"`
	Description            string       `gorm:"index"`
	WorkflowId             string
	Status                 domain.ClusterStatus
	StatusDesc             string
	CloudAccountId         *uuid.UUID
	CloudAccount           CloudAccount `gorm:"foreignKey:CloudAccountId"`
	StackTemplateId        uuid.UUID
	StackTemplate          StackTemplate `gorm:"foreignKey:StackTemplateId"`
	Favorites              *[]ClusterFavorite
	ClusterType            domain.ClusterType `gorm:"default:0"`
	ByoClusterEndpointHost string
	ByoClusterEndpointPort int
	IsStack                bool `gorm:"default:false"`
	TksCpNode              int
	TksCpNodeMax           int
	TksCpNodeType          string
	TksInfraNode           int
	TksInfraNodeMax        int
	TksInfraNodeType       string
	TksUserNode            int
	TksUserNodeMax         int
	TksUserNodeType        string
	Kubeconfig             []byte     `gorm:"-:all"`
	PolicyIds              []string   `gorm:"-:all"`
	CreatorId              *uuid.UUID `gorm:"type:uuid"`
	Creator                User       `gorm:"foreignKey:CreatorId"`
	UpdatorId              *uuid.UUID `gorm:"type:uuid"`
	Updator                User       `gorm:"foreignKey:UpdatorId"`
}

func (m *Cluster) SetDefaultConf() {
	m.TksCpNodeMax = m.TksCpNode

	if m.TksInfraNode == 0 {
		m.TksInfraNode = 3
	}
	m.TksInfraNodeMax = m.TksInfraNode

	if m.TksUserNode == 0 {
		m.TksUserNode = 1
	}
	m.TksUserNodeMax = m.TksUserNode

	if m.TksCpNodeType == "" {
		m.TksCpNodeType = "t3.xlarge"
	}
	if m.TksInfraNodeType == "" {
		m.TksInfraNodeType = "t3.2xlarge"
	}
	if m.TksUserNodeType == "" {
		m.TksUserNodeType = "t3.large"
	}
}

type ClusterFavorite struct {
	gorm.Model

	ID        uuid.UUID `gorm:"primarykey;type:uuid"`
	ClusterId domain.ClusterId
	Cluster   Cluster   `gorm:"foreignKey:ClusterId"`
	UserId    uuid.UUID `gorm:"type:uuid"`
	User      User      `gorm:"foreignKey:UserId"`
}
