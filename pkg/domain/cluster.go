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
	ID             ClusterId
	OrganizationId string
	Name           string
	Description    string
	CloudSettingId uuid.UUID
	CloudSetting   CloudSetting
	Status         ClusterStatus
	StatusDesc     string
	Conf           ClusterConf
	TemplateId     string
	CreatorId      *uuid.UUID
	Creator        User
	UpdatorId      *uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ClusterConf = struct {
	SshKeyName      string
	Region          string
	MachineType     string
	NumOfAz         int
	MinSizePerAz    int
	MaxSizePerAz    int
	MachineReplicas int
}

type ClusterCapacity = struct {
	Max     int
	Current int
}

type ClusterKubeInfo = struct {
	Version        string          `json:"version"`
	TotalResources int             `json:"totalResources"`
	Nodes          int             `json:"nodes"`
	Namespaces     int             `json:"namespaces"`
	Services       int             `json:"services"`
	Pods           int             `json:"pods"`
	Cores          ClusterCapacity `json:"cores"`
	Memory         ClusterCapacity `json:"memory"`
	Updated        time.Time       `json:"updated"`
}

type Event = struct {
	ID        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Updated   time.Time `json:"updated"`
}

type Node = struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	InstanceType string    `json:"instanceType"`
	Role         string    `json:"role"`
	Updated      time.Time `json:"updated"`
}

type CreateClusterRequest struct {
	OrganizationId  string `json:"organizationId"`
	TemplateId      string `json:"templateId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CloudSettingId  string `json:"cloudSettingId"`
	NumOfAz         int    `json:"numOfAz"`
	MachineType     string `json:"machineType"`
	Region          string `json:"region"`
	MachineReplicas int    `json:"machineReplicas"`
}

type CreateClusterResponse struct {
	ID string `json:"id"`
}

type ClusterResponse struct {
	ID             ClusterId            `json:"id"`
	OrganizationId string               `json:"organizationId"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	CloudSetting   CloudSettingResponse `json:"cloudSetting"`
	Status         string               `json:"status"`
	StatusDesc     string               `json:"statusDesc"`
	Conf           ClusterConfResponse  `json:"conf"`
	Creator        SimpleUserResponse   `json:"creator"`
	Updator        SimpleUserResponse   `json:"updator"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

type ClusterConfResponse struct {
	SshKeyName      string `json:"sshKeyName"`
	Region          string `json:"region"`
	MachineType     string `json:"machineType"`
	NumOfAz         int    `json:"numOfAz"`
	MinSizePerAz    int    `json:"minSizePerAz"`
	MaxSizePerAz    int    `json:"maxSizePerAz"`
	MachineReplicas int    `json:"machineReplicas"`
}

type GetClustersResponse struct {
	Clusters []ClusterResponse `json:"clusters"`
}

type GetClusterResponse struct {
	Cluster ClusterResponse `json:"cluster"`
}
