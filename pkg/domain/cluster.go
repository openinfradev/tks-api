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
type Cluster = struct {
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

type ClusterConf = struct {
	Region              string
	CpNodeCnt           string
	CpNodeMachineType   string
	TksNodeCnt          string
	TksNodeMachineType  string
	UserNodeCnt         string
	UserNodeMachineType string
}

type CreateClusterRequest struct {
	OrganizationId      string `json:"organizationId"`
	StackTemplateId     string `json:"stackTemplateId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	CloudAccountId      string `json:"cloudAccountId"`
	Region              string `json:"region"`
	CpNodeCnt           string `json:"cpNodeCnt"`
	CpNodeMachineType   string `json:"cpNodeMachineType"`
	TksNodeCnt          string `json:"tksNodeCnt"`
	TksNodeMachineType  string `json:"tksNodeMachineType"`
	UserNodeCnt         string `json:"userNodeCnt"`
	UserNodeMachineType string `json:"userNodeMachineType"`
}

type CreateClusterResponse struct {
	ID string `json:"id"`
}

type ClusterResponse struct {
	ID             ClusterId             `json:"id"`
	OrganizationId string                `json:"organizationId"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	CloudAccount   CloudAccountResponse  `json:"cloudAccount"`
	StackTemplate  StackTemplateResponse `json:"stackTemplate"`
	Status         string                `json:"status"`
	StatusDesc     string                `json:"statusDesc"`
	Conf           ClusterConfResponse   `json:"conf"`
	Creator        SimpleUserResponse    `json:"creator"`
	Updator        SimpleUserResponse    `json:"updator"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type ClusterConfResponse struct {
	CpNodeCnt   int `json:"cpNodeCnt"`
	TksNodeCnt  int `json:"tksNodeCnt"`
	UserNodeCnt int `json:"userpNodeCnt"`
}

type GetClustersResponse struct {
	Clusters []ClusterResponse `json:"clusters"`
}

type GetClusterResponse struct {
	Cluster ClusterResponse `json:"cluster"`
}
