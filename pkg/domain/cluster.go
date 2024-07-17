package domain

import (
	"time"

	"github.com/openinfradev/tks-api/internal/helper"
)

const NODE_TYPE_TKS_CP_NODE = "TKS_CP_NODE"
const NODE_TYPE_TKS_INFRA_NODE = "TKS_INFRA_NODE"
const NODE_TYPE_TKS_USER_NODE = "TKS_USER_NODE"
const MAX_AZ_NUM = 3

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
	ClusterStatus_BOOTSTRAPPING
	ClusterStatus_BOOTSTRAPPED
	ClusterStatus_BOOTSTRAP_ERROR
	ClusterStatus_STOPPED
)

var clusterStatus = [...]string{
	"PENDING",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
	"INSTALL_ERROR",
	"DELETE_ERROR",
	"BOOTSTRAPPING",
	"BOOTSTRAPPED",
	"BOOTSTRAP_ERROR",
	"STOPPED",
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

type ClusterType int32

const (
	ClusterType_USER = iota
	ClusterType_ADMIN
)

var clusterType = [...]string{
	"USER",
	"ADMIN",
}

func (m ClusterType) String() string { return clusterType[(m)] }
func (m ClusterType) FromString(s string) ClusterType {
	for i, v := range clusterType {
		if v == s {
			return ClusterType(i)
		}
	}
	return ClusterType_USER
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

type ClusterHost struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ClusterNode struct {
	Type        string        `json:"type"`
	Targeted    int           `json:"targeted"`
	Registered  int           `json:"registered"`
	Registering int           `json:"registering"`
	Status      string        `json:"status"`
	Command     string        `json:"command"`
	Validity    int           `json:"validity"`
	Hosts       []ClusterHost `json:"hosts"`
}

type BootstrapKubeconfig struct {
	Expiration int `json:"expiration"`
}

type ClusterDomain struct {
	DomainType string `json:"domainType" validate:"oneof=grafana loki minio thanos_sidecar jaeger kiali"`
	Url        string `json:"url"`
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
	OrganizationId         string          `json:"organizationId" validate:"required"`
	CloudService           string          `json:"cloudService" validate:"required,oneof=AWS BYOH"`
	StackTemplateId        string          `json:"stackTemplateId" validate:"required"`
	Name                   string          `json:"name" validate:"required,name"`
	Description            string          `json:"description"`
	CloudAccountId         string          `json:"cloudAccountId"`
	ClusterType            string          `json:"clusterType"`
	ByoClusterEndpointHost string          `json:"byoClusterEndpointHost,omitempty"`
	ByoClusterEndpointPort int             `json:"byoClusterEndpointPort,omitempty"`
	IsStack                bool            `json:"isStack,omitempty"`
	PolicyIds              []string        `json:"policyIds,omitempty"`
	TksCpNode              int             `json:"tksCpNode"`
	TksCpNodeMax           int             `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType          string          `json:"tksCpNodeType,omitempty"`
	TksInfraNode           int             `json:"tksInfraNode"`
	TksInfraNodeMax        int             `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType       string          `json:"tksInfraNodeType,omitempty"`
	TksUserNode            int             `json:"tksUserNode"`
	TksUserNodeMax         int             `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType        string          `json:"tksUserNodeType,omitempty"`
	Domains                []ClusterDomain `json:"domains,omitempty"`
}

type ImportClusterRequest struct {
	OrganizationId  string          `json:"organizationId" validate:"required"`
	StackTemplateId string          `json:"stackTemplateId" validate:"required"`
	Name            string          `json:"name" validate:"required,name"`
	Description     string          `json:"description"`
	ClusterType     string          `json:"clusterType"`
	Kubeconfig      []byte          `json:"kubeconfig"`
	CloudService    string          `json:"cloudService"`
	PolicyIds       []string        `json:"policyIds,omitempty"`
	Domains         []ClusterDomain `json:"domains,omitempty"`
}

type CreateClusterResponse struct {
	ID string `json:"id"`
}

type ImportClusterResponse struct {
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
	ID                     ClusterId                   `json:"id"`
	CloudService           string                      `json:"cloudService"`
	OrganizationId         string                      `json:"organizationId"`
	Name                   string                      `json:"name"`
	Description            string                      `json:"description"`
	CloudAccount           SimpleCloudAccountResponse  `json:"cloudAccount"`
	StackTemplate          SimpleStackTemplateResponse `json:"stackTemplate"`
	Status                 string                      `json:"status"`
	StatusDesc             string                      `json:"statusDesc"`
	Conf                   ClusterConfResponse         `json:"conf"`
	ClusterType            string                      `json:"clusterType"`
	Creator                SimpleUserResponse          `json:"creator"`
	Updator                SimpleUserResponse          `json:"updator"`
	CreatedAt              time.Time                   `json:"createdAt"`
	UpdatedAt              time.Time                   `json:"updatedAt"`
	ByoClusterEndpointHost string                      `json:"byoClusterEndpointHost,omitempty"`
	ByoClusterEndpointInt  int                         `json:"byoClusterEndpointPort,omitempty"`
	IsStack                bool                        `json:"isStack,omitempty"`
	Favorited              bool                        `json:"favorited,omitempty"`
	Domains                []ClusterDomain             `json:"domains"`
}

type SimpleClusterResponse struct {
	ID             ClusterId `json:"id"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
}

type ClusterSiteValuesResponse struct {
	ClusterType            string          `json:"clusterType"`
	SshKeyName             string          `json:"sshKeyName"`
	ClusterRegion          string          `json:"clusterRegion"`
	TksCpNode              int             `json:"tksCpNode"`
	TksCpNodeMax           int             `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType          string          `json:"tksCpNodeType,omitempty"`
	TksInfraNode           int             `json:"tksInfraNode"`
	TksInfraNodeMax        int             `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType       string          `json:"tksInfraNodeType,omitempty"`
	TksUserNode            int             `json:"tksUserNode"`
	TksUserNodeMax         int             `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType        string          `json:"tksUserNodeType,omitempty"`
	ByoClusterEndpointHost string          `json:"byoClusterEndpointHost"`
	ByoClusterEndpointPort int             `json:"byoClusterEndpointPort"`
	Domains                []ClusterDomain `json:"domains,omitempty"`
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

type InstallClusterRequest struct {
	ClusterId      string `json:"clusterId" validate:"required"`
	OrganizationId string `json:"organizationId" validate:"required"`
}

type CreateBootstrapKubeconfigResponse struct {
	Data BootstrapKubeconfig `json:"kubeconfig"`
}

type GetBootstrapKubeconfigResponse struct {
	Data BootstrapKubeconfig `json:"kubeconfig"`
}

type GetClusterNodesResponse struct {
	Nodes []ClusterNode `json:"nodes"`
}
