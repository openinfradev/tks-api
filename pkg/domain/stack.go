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
	StackStatus_INSTALLING
	StackStatus_RUNNING
	StackStatus_DELETING
	StackStatus_DELETED
	StackStatus_ERROR
)

var stackStatus = [...]string{
	"PENDING",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
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
	CreatorId       *uuid.UUID
	Creator         User
	UpdatorId       *uuid.UUID
	Updator         User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StackConf = struct {
	CpNodeCnt           int
	CpNodeMachineType   string
	TksNodeCnt          int
	TksNodeMachineType  string
	UserNodeCnt         int
	UserNodeMachineType string
}

type CreateStackRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	StackTemplateId     string `json:"stackTemplateId"`
	CloudAccountId      string `json:"cloudAccountId"`
	CpNodeCnt           int    `json:"cpNodeCnt"`
	CpNodeMachineType   string `json:"cpNodeMachineType"`
	TksNodeCnt          int    `json:"tksNodeCnt"`
	TksNodeMachineType  string `json:"tksNodeMachineType"`
	UserNodeCnt         int    `json:"userNodeCnt"`
	UserNodeMachineType string `json:"userNodeMachineType"`
}

type CreateStackResponse struct {
	ID string `json:"id"`
}

type StackResponse struct {
	ID             StackId               `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	OrganizationId string                `json:"organizationId"`
	StackTemplate  StackTemplateResponse `json:"stackTemplate,omitempty"`
	CloudAccount   CloudAccountResponse  `json:"cloudAccount,omitempty"`
	Status         string                `json:"status"`
	StatusDesc     string                `json:"statusDesc"`
	PrimaryCluster bool                  `json:"primaryCluster"`
	Conf           StackConf             `json:"conf"`
	Creator        SimpleUserResponse    `json:"creator,omitempty"`
	Updator        SimpleUserResponse    `json:"updator,omitempty"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type GetStacksResponse struct {
	Stacks []StackResponse `json:"stacks"`
}

type GetStackResponse struct {
	Stack StackResponse `json:"stack"`
}

type UpdateStackRequest struct {
	ID StackId `json:"id"`
}

type DeleteStackRequest struct {
	ID StackId `json:"id"`
}

type CheckStackNameResponse struct {
	Existed bool `json:"existed"`
}
