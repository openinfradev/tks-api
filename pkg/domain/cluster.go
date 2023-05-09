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
	ClusterStatus_ERROR
)

var clusterStatus = [...]string{
	"PENDING",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
	"ERROR",
}

func (m ClusterStatus) String() string { return clusterStatus[(m)] }
func (m ClusterStatus) FromString(s string) ClusterStatus {
	for i, v := range clusterStatus {
		if v == s {
			return ClusterStatus(i)
		}
	}
	return ClusterStatus_ERROR
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
	CpNodeCnt           int
	CpNodeMachineType   string
	TksNodeCnt          int
	TksNodeMachineType  string
	UserNodeCnt         int
	UserNodeMachineType string
}

// [TODO] annotaion 으로 가능하려나?
func (m *ClusterConf) SetDefault() {
	if m.CpNodeCnt == 0 {
		m.CpNodeCnt = 3
	}
	if m.TksNodeCnt == 0 {
		m.TksNodeCnt = 3
	}
	if m.UserNodeCnt == 0 {
		m.UserNodeCnt = 1
	}
	if m.CpNodeMachineType == "" {
		m.CpNodeMachineType = "t3.xlarge"
	}
	if m.TksNodeMachineType == "" {
		m.TksNodeMachineType = "t3.2xlarge"
	}
	if m.UserNodeMachineType == "" {
		m.UserNodeMachineType = "t3.large"
	}
}

type CreateClusterRequest struct {
	OrganizationId      string `json:"organizationId" validate:"required"`
	StackTemplateId     string `json:"stackTemplateId" validate:"required"`
	Name                string `json:"name" validate:"required"`
	Description         string `json:"description"`
	CloudAccountId      string `json:"cloudAccountId" validate:"required"`
	CpNodeCnt           int    `json:"cpNodeCnt,omitempty"`
	CpNodeMachineType   string `json:"cpNodeMachineType,omitempty"`
	TksNodeCnt          int    `json:"tksNodeCnt",omitempty`
	TksNodeMachineType  string `json:"tksNodeMachineType,omitempty"`
	UserNodeCnt         int    `json:"userNodeCnt",omitempty`
	UserNodeMachineType string `json:"userNodeMachineType,omitempty"`
}

type CreateClusterResponse struct {
	ID string `json:"id"`
}

type ClusterConfResponse struct {
	CpNodeCnt   int `json:"cpNodeCnt"`
	TksNodeCnt  int `json:"tksNodeCnt"`
	UserNodeCnt int `json:"userpNodeCnt"`
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
	SshKeyName        string `json:"sshKeyName"`
	ClusterRegion     string `json:"clusterRegion"`
	CpReplicas        int    `json:"cpReplicas"`
	CpNodeMachineType string `json:"cpNodeMachineType"`
	MpReplicas        int    `json:"mpReplicas"`
	MpNodeMachineType string `json:"mpNodeMachineType"`
	MdNumOfAz         int    `json:"mdNumOfAz"`
	MdMinSizePerAz    int    `json:"mdMinSizePerAz"`
	MdMaxSizePerAz    int    `json:"mdMaxSizePerAz"`
	MdMachineType     string `json:"mdMachineType"`
}

type GetClustersResponse struct {
	Clusters []ClusterResponse `json:"clusters"`
}

type GetClusterResponse struct {
	Cluster ClusterResponse `json:"cluster"`
}

type GetClusterSiteValuesResponse struct {
	ClusterSiteValues ClusterSiteValuesResponse `json:"clusterSiteValues"`
}
