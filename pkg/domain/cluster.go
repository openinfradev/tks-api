package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
)

type ClusterId string

func (c ClusterId) String() string {
	return string(c)
}

func (c ClusterId) Validate() bool {
	return helper.ValidateClusterId(c.String())
}

// enum
type ClusterStatus int32

const (
	ClusterStatus_PENDING ClusterStatus = iota
	ClusterStatus_INSTALLING
	ClusterStatus_RUNNING
	ClusterStatus_DELETING
	ClusterStatus_DELETED
	ClusterStatus_INSTALL_ERROR
	ClusterStatus_DELETE_ERROR
)

var clusterStatus = [...]string{
	"PENDING",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
	"INSTALL_ERROR",
	"DELETE_ERROR",
}

func (m ClusterStatus) String() string { return clusterStatus[(m)] }
func (m ClusterStatus) FromString(s string) ClusterStatus {
	for i, v := range clusterStatus {
		if v == s {
			return ClusterStatus(i)
		}
	}
	return ClusterStatus_PENDING
}

// model
type Cluster struct {
	ID              ClusterId
	OrganizationId  string
	Name            string
	Description     string
	CloudAccountId  uuid.UUID
	CloudAccount    CloudAccount
	StackTemplateId uuid.UUID
	StackTemplate   StackTemplate
	Status          ClusterStatus
	StatusDesc      string
	Conf            ClusterConf
	CreatorId       *uuid.UUID
	Creator         User
	UpdatorId       *uuid.UUID
	Updator         User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ClusterConf struct {
	TksCpNode        int
	TksCpNodeMax     int
	TksCpNodeType    string
	TksInfraNode     int
	TksInfraNodeMax  int
	TksInfraNodeType string
	TksUserNode      int
	TksUserNodeMax   int
	TksUserNodeType  string
}

// [TODO] annotaion 으로 가능하려나?
func (m *ClusterConf) SetDefault() {
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

type CreateClusterRequest struct {
	OrganizationId   string `json:"organizationId" validate:"required"`
	StackTemplateId  string `json:"stackTemplateId" validate:"required"`
	Name             string `json:"name" validate:"required,name"`
	Description      string `json:"description"`
	CloudAccountId   string `json:"cloudAccountId" validate:"required"`
	TksCpNode        int    `json:"tksCpNode"`
	TksCpNodeMax     int    `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType    string `json:"tksCpNodeType,omitempty"`
	TksInfraNode     int    `json:"tksInfraNode"`
	TksInfraNodeMax  int    `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType string `json:"tksInfraNodeType,omitempty"`
	TksUserNode      int    `json:"tksUserNode"`
	TksUserNodeMax   int    `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType  string `json:"tksUserNodeType,omitempty"`
}

type CreateClusterResponse struct {
	ID string `json:"id"`
}

type ClusterConfResponse struct {
	TksCpNode        int    `json:"tksCpNode"`
	TksCpNodeMax     int    `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType    string `json:"tksCpNodeType,omitempty"`
	TksInfraNode     int    `json:"tksInfraNode"`
	TksInfraNodeMax  int    `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType string `json:"tksInfraNodeType,omitempty"`
	TksUserNode      int    `json:"tksUserNode"`
	TksUserNodeMax   int    `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType  string `json:"tksUserNodeType,omitempty"`
}

type ClusterResponse struct {
	ID             ClusterId                   `json:"id"`
	OrganizationId string                      `json:"organizationId"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	CloudAccount   SimpleCloudAccountResponse  `json:"cloudAccount"`
	StackTemplate  SimpleStackTemplateResponse `json:"stackTemplate"`
	Status         string                      `json:"status"`
	StatusDesc     string                      `json:"statusDesc"`
	Conf           ClusterConfResponse         `json:"conf"`
	Creator        SimpleUserResponse          `json:"creator"`
	Updator        SimpleUserResponse          `json:"updator"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

type SimpleClusterResponse struct {
	ID             ClusterId `json:"id"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
}

type ClusterSiteValuesResponse struct {
	SshKeyName       string `json:"sshKeyName"`
	ClusterRegion    string `json:"clusterRegion"`
	TksCpNode        int    `json:"tksCpNode"`
	TksCpNodeMax     int    `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType    string `json:"tksCpNodeType,omitempty"`
	TksInfraNode     int    `json:"tksInfraNode"`
	TksInfraNodeMax  int    `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType string `json:"tksInfraNodeType,omitempty"`
	TksUserNode      int    `json:"tksUserNode"`
	TksUserNodeMax   int    `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType  string `json:"tksUserNodeType,omitempty"`
}

type GetClustersResponse struct {
	Clusters   []ClusterResponse  `json:"clusters"`
	Pagination PaginationResponse `json:"pagination"`
}

type GetClusterResponse struct {
	Cluster ClusterResponse `json:"cluster"`
}

type GetClusterSiteValuesResponse struct {
	ClusterSiteValues ClusterSiteValuesResponse `json:"clusterSiteValues"`
}
