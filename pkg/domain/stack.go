package domain

import (
	"time"

	"github.com/google/uuid"
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
	StackStatus_APPGROUP_ERROR

	StackStatus_CLUSTER_INSTALLING
	StackStatus_CLUSTER_DELETING
	StackStatus_CLUSTER_DELETED
	StackStatus_CLUSTER_ERROR

	StackStatus_RUNNING
	StackStatus_ERROR
)

var stackStatus = [...]string{
	"PENDING",
	"APPGROUP_INSTALLING",
	"APPGROUP_DELETING",
	"APPGROUP_ERROR",
	"CLUSTER_INSTALLING",
	"CLUSTER_DELETING",
	"CLUSTER_DELETED",
	"CLUSTER_ERROR",
	"RUNNING",
	"ERROR",
}

func (m StackStatus) String() string { return stackStatus[(m)] }
func (m StackStatus) FromString(s string) StackStatus {
	for i, v := range stackStatus {
		if v == s {
			return StackStatus(i)
		}
	}
	return StackStatus_ERROR
}

const MAX_STEP_CLUSTER_CREATE = 14
const MAX_STEP_CLUSTER_REMOVE = 11
const MAX_STEP_LMA_CREATE_PRIMARY = 36
const MAX_STEP_LMA_CREATE_MEMBER = 27
const MAX_STEP_LMA_REMOVE = 9
const MAX_STEP_SM_CREATE = 22
const MAX_STEP_SM_REMOVE = 4

// model
type Stack = struct {
	ID              StackId
	Name            string
	Description     string
	OrganizationId  string
	CloudAccountId  uuid.UUID
	CloudAccount    CloudAccount
	StackTemplateId uuid.UUID
	StackTemplate   StackTemplate
	Status          StackStatus
	StatusDesc      string
	Conf            StackConf
	PrimaryCluster  bool
	GrafanaUrl      string
	CreatorId       *uuid.UUID
	Creator         User
	UpdatorId       *uuid.UUID
	Updator         User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StackConf struct {
	CpNodeCnt           int
	CpNodeMachineType   string
	TksNodeCnt          int
	TksNodeMachineType  string
	UserNodeCnt         int
	UserNodeMachineType string
}

type StackStepStatus struct {
	Status  string `json:"status"`
	Stage   string `json:"stage"`
	Step    int    `json:"step"`
	MaxStep int    `json:"maxStep"`
}

type CreateStackRequest struct {
	Name                string `json:"name" validate:"required,name"`
	Description         string `json:"description"`
	StackTemplateId     string `json:"stackTemplateId" validate:"required"`
	CloudAccountId      string `json:"cloudAccountId" validate:"required"`
	CpNodeCnt           int    `json:"cpNodeCnt,omitempty"`
	CpNodeMachineType   string `json:"cpNodeMachineType,omitempty"`
	TksNodeCnt          int    `json:"tksNodeCnt" validate:"required,min=3,max=6"`
	TksNodeMachineType  string `json:"tksNodeMachineType,omitempty"`
	UserNodeCnt         int    `json:"userNodeCnt" validate:"required,min=0,max=100"`
	UserNodeMachineType string `json:"userNodeMachineType,omitempty"`
}

type CreateStackResponse struct {
	ID string `json:"id"`
}

type StackConfResponse struct {
	CpNodeCnt           int    `json:"cpNodeCnt"`
	CpNodeMachineType   string `json:"cpNodeMachineType,omitempty"`
	TksNodeCnt          int    `json:"tksNodeCnt"`
	TksNodeMachineType  string `json:"tksNodeMachineType,omitempty"`
	UserNodeCnt         int    `json:"userNodeCnt"`
	UserNodeMachineType string `json:"userNodeMachineType,omitempty"`
}

type StackResponse struct {
	ID             StackId                     `json:"id"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	OrganizationId string                      `json:"organizationId"`
	StackTemplate  SimpleStackTemplateResponse `json:"stackTemplate,omitempty"`
	CloudAccount   SimpleCloudAccountResponse  `json:"cloudAccount,omitempty"`
	Status         string                      `json:"status"`
	StatusDesc     string                      `json:"statusDesc"`
	PrimaryCluster bool                        `json:"primaryCluster"`
	Conf           StackConfResponse           `json:"conf"`
	GrafanaUrl     string                      `json:"grafanaUrl"`
	Creator        SimpleUserResponse          `json:"creator,omitempty"`
	Updator        SimpleUserResponse          `json:"updator,omitempty"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

type GetStacksResponse struct {
	Stacks []StackResponse `json:"stacks"`
}

type GetStackResponse struct {
	Stack StackResponse `json:"stack"`
}

type UpdateStackRequest struct {
	Description string `json:"description" validate:"required"`
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
