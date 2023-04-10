package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
)

type StackId ClusterId

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
	ID         StackId
	Cluster    Cluster
	AppGroups  []AppGroup
	Status     StackStatus
	StatusDesc string
	CreatorId  *uuid.UUID
	Creator    User
	UpdatorId  *uuid.UUID
	Updator    User
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateStackRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	OrganizationId  string `json:"organizationId"`
	StackTemplateId string `json:"stackTemplateId"`
	CloudSettingId  string `json:"cloudSettingId"`
	NumOfAz         int    `json:"numOfAz"`
	MachineType     string `json:"machineType"`
	Region          string `json:"region"`
	MachineReplicas int    `json:"machineReplicas"`
}

type CreateStackResponse struct {
	ID string `json:"id"`
}

type StackResponse struct {
	ID             StackId               `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	OrganizationId string                `json:"organizationId"`
	StackTemplate  StackTemplateResponse `json:"stackTemplate"`
	CloudSetting   CloudSettingResponse  `json:"cloudSetting"`
	Status         string                `json:"status"`
	StatusDesc     string                `json:"statusDesc"`
	Creator        SimpleUserResponse    `json:"creator"`
	Updator        SimpleUserResponse    `json:"updator"`
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
