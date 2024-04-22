package domain

import (
	"time"

	"github.com/openinfradev/tks-api/internal/helper"
)

type StackId string

func (c StackId) String() string {
	return string(c)
}

func (c StackId) Validate() bool {
	return helper.ValidateClusterId(c.String())
}

// enum
type StackStatus int32

const (
	StackStatus_PENDING StackStatus = iota

	StackStatus_APPGROUP_INSTALLING
	StackStatus_APPGROUP_DELETING
	StackStatus_APPGROUP_INSTALL_ERROR
	StackStatus_APPGROUP_DELETE_ERROR

	StackStatus_CLUSTER_INSTALLING
	StackStatus_CLUSTER_DELETING
	StackStatus_CLUSTER_DELETED
	StackStatus_CLUSTER_INSTALL_ERROR
	StackStatus_CLUSTER_DELETE_ERROR

	StackStatus_RUNNING

	StackStatus_CLUSTER_BOOTSTRAPPING
	StackStatus_CLUSTER_BOOTSTRAPPED
)

var stackStatus = [...]string{
	"PENDING",
	"APPGROUP_INSTALLING",
	"APPGROUP_DELETING",
	"APPGROUP_INSTALL_ERROR",
	"APPGROUP_DELETE_ERROR",
	"CLUSTER_INSTALLING",
	"CLUSTER_DELETING",
	"CLUSTER_DELETED",
	"CLUSTER_INSTALL_ERROR",
	"CLUSTER_DELETE_ERROR",
	"RUNNING",
	"BOOTSTRAPPING",
	"BOOTSTRAPPED",
}

func (m StackStatus) String() string { return stackStatus[(m)] }
func (m StackStatus) FromString(s string) StackStatus {
	for i, v := range stackStatus {
		if v == s {
			return StackStatus(i)
		}
	}
	return StackStatus_PENDING
}

const MAX_STEP_CLUSTER_CREATE = 26
const MAX_STEP_CLUSTER_REMOVE = 16
const MAX_STEP_LMA_CREATE_PRIMARY = 39
const MAX_STEP_LMA_CREATE_MEMBER = 29
const MAX_STEP_LMA_REMOVE = 12
const MAX_STEP_SM_CREATE = 22
const MAX_STEP_SM_REMOVE = 4

type StackStepStatus struct {
	Status  string `json:"status"`
	Stage   string `json:"stage"`
	Step    int    `json:"step"`
	MaxStep int    `json:"maxStep"`
}

type CreateStackRequest struct {
	Name             string   `json:"name" validate:"required,name,rfc1123"`
	Description      string   `json:"description"`
	ClusterId        string   `json:"clusterId"`
	CloudService     string   `json:"cloudService" validate:"required,oneof=AWS BYOH"`
	StackTemplateId  string   `json:"stackTemplateId" validate:"required"`
	CloudAccountId   string   `json:"cloudAccountId"`
	ClusterEndpoint  string   `json:"userClusterEndpoint,omitempty"`
	PolicyIds        []string `json:"policyIds,omitempty"`
	TksCpNode        int      `json:"tksCpNode"`
	TksCpNodeMax     int      `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType    string   `json:"tksCpNodeType,omitempty"`
	TksInfraNode     int      `json:"tksInfraNode"`
	TksInfraNodeMax  int      `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType string   `json:"tksInfraNodeType,omitempty"`
	TksUserNode      int      `json:"tksUserNode"`
	TksUserNodeMax   int      `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType  string   `json:"tksUserNodeType,omitempty"`
}

type CreateStackResponse struct {
	ID string `json:"id"`
}

type StackConfResponse struct {
	TksCpNode        int    `json:"tksCpNode"`
	TksCpNodeMax     int    `json:"tksCpNodeMax,omitempty"`
	TksCpNodeType    string `json:"tksCpNodeType,omitempty"`
	TksInfraNode     int    `json:"tksInfraNode" validate:"required,min=1,max=3"`
	TksInfraNodeMax  int    `json:"tksInfraNodeMax,omitempty"`
	TksInfraNodeType string `json:"tksInfraNodeType,omitempty"`
	TksUserNode      int    `json:"tksUserNode" validate:"required,min=0,max=100"`
	TksUserNodeMax   int    `json:"tksUserNodeMax,omitempty"`
	TksUserNodeType  string `json:"tksUserNodeType,omitempty"`
}

type StackResponse struct {
	ID              StackId                     `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	OrganizationId  string                      `json:"organizationId"`
	StackTemplate   SimpleStackTemplateResponse `json:"stackTemplate,omitempty"`
	CloudAccount    SimpleCloudAccountResponse  `json:"cloudAccount,omitempty"`
	Status          string                      `json:"status"`
	StatusDesc      string                      `json:"statusDesc"`
	PrimaryCluster  bool                        `json:"primaryCluster"`
	Conf            StackConfResponse           `json:"conf"`
	GrafanaUrl      string                      `json:"grafanaUrl"`
	Creator         SimpleUserResponse          `json:"creator,omitempty"`
	Updator         SimpleUserResponse          `json:"updator,omitempty"`
	Favorited       bool                        `json:"favorited"`
	ClusterEndpoint string                      `json:"userClusterEndpoint,omitempty"`
	Resource        DashboardStackResponse      `json:"resource,omitempty"`
	CreatedAt       time.Time                   `json:"createdAt"`
	UpdatedAt       time.Time                   `json:"updatedAt"`
}

type SimpleStackResponse struct {
	ID             StackId                     `json:"id"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	OrganizationId string                      `json:"organizationId"`
	StackTemplate  SimpleStackTemplateResponse `json:"stackTemplate,omitempty"`
	CloudAccount   SimpleCloudAccountResponse  `json:"cloudAccount,omitempty"`
	Status         string                      `json:"status"`
	PrimaryCluster bool                        `json:"primaryCluster"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

type GetStacksResponse struct {
	Stacks     []StackResponse    `json:"stacks"`
	Pagination PaginationResponse `json:"pagination"`
}

type GetStackResponse struct {
	Stack StackResponse `json:"stack"`
}

type UpdateStackRequest struct {
	Description string `json:"description"`
}

type CheckStackNameResponse struct {
	Existed bool `json:"existed"`
}

type GetStackKubeConfigResponse struct {
	KubeConfig string `json:"kubeConfig"`
}

type GetStackStatusResponse struct {
	StackStatus string            `json:"stackStatus"`
	StepStatus  []StackStepStatus `json:"stepStatus"`
}
