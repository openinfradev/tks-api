package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
)

type AppGroupId string

func (c AppGroupId) String() string {
	return string(c)
}

func (c AppGroupId) Validate() bool {
	return helper.ValidateApplicationGroupId(c.String())
}

// enum
type AppGroupStatus int32

const (
	AppGroupStatus_PENDING AppGroupStatus = iota
	AppGroupStatus_INSTALLING
	AppGroupStatus_RUNNING
	AppGroupStatus_DELETING
	AppGroupStatus_DELETED
	AppGroupStatus_ERROR
)

var appGroupStatus = [...]string{
	"PENDING",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
	"ERROR",
}

func (m AppGroupStatus) String() string { return appGroupStatus[(m)] }
func (m AppGroupStatus) FromString(s string) AppGroupStatus {
	for i, v := range appGroupStatus {
		if v == s {
			return AppGroupStatus(i)
		}
	}
	return AppGroupStatus_PENDING
}

type ApplicationType int32

const (
	ApplicationType_UNSPECIFIED ApplicationType = iota
	ApplicationType_THANOS
	ApplicationType_PROMETHEUS
	ApplicationType_GRAFANA
	ApplicationType_KIALI
	ApplicationType_KIBANA
	ApplicationType_ELASTICSERCH
	ApplicationType_CLOUD_CONSOLE
	ApplicationType_HORIZON
	ApplicationType_JAEGER
	ApplicationType_KUBERNETES_DASHBOARD
)

var applicationType = [...]string{
	"UNSPECIFIED",
	"THANOS",
	"PROMETHEUS",
	"GRAFANA",
	"KIALI",
	"KIBANA",
	"ELASTICSERCH",
	"CLOUD_CONSOLE",
	"HORIZON",
	"JAEGER",
	"KUBERNETES_DASHBOARD",
}

func (m ApplicationType) String() string { return applicationType[(m)] }
func (m ApplicationType) FromString(s string) ApplicationType {
	for i, v := range applicationType {
		if v == s {
			return ApplicationType(i)
		}
	}
	return ApplicationType_UNSPECIFIED
}

type AppGroupType int32

const (
	AppGroupType_UNSPECIFIED AppGroupType = iota
	AppGroupType_LMA
	AppGroupType_SERVICE_MESH
)

var appGroupType = [...]string{
	"UNSPECIFIED",
	"LMA",
	"SERVICE_MESH",
}

func (m AppGroupType) String() string { return appGroupType[(m)] }
func (m AppGroupType) FromString(s string) AppGroupType {
	for i, v := range appGroupType {
		if v == s {
			return AppGroupType(i)
		}
	}
	return AppGroupType_UNSPECIFIED
}

type AppGroup = struct {
	ID           AppGroupId
	Name         string
	ClusterId    ClusterId
	AppGroupType AppGroupType
	Description  string
	WorkflowId   string
	Status       AppGroupStatus
	StatusDesc   string
	CreatorId    *uuid.UUID
	Creator      User
	UpdatorId    *uuid.UUID
	Updator      User
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Application = struct {
	ID              uuid.UUID
	AppGroupId      AppGroupId
	Endpoint        string
	Metadata        string
	ApplicationType ApplicationType
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AppGroupResponse = struct {
	ID           AppGroupId         `json:"id"`
	Name         string             `json:"name"`
	ClusterId    ClusterId          `json:"clusterId"`
	AppGroupType AppGroupType       `json:"appGroupType"`
	Description  string             `json:"description"`
	WorkflowId   string             `json:"workflowId"`
	Status       AppGroupStatus     `json:"status"`
	StatusDesc   string             `json:"statusDesc"`
	Creator      SimpleUserResponse `json:"creator"`
	Updator      SimpleUserResponse `json:"updator"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

type ApplicationResponse = struct {
	ID              uuid.UUID       `json:"id"`
	AppGroupId      AppGroupId      `json:"appGroupId"`
	Endpoint        string          `json:"endpoint"`
	Metadata        string          `json:"metadata"`
	ApplicationType ApplicationType `json:"applicationType"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type CreateAppGroupRequest struct {
	Name         string    `json:"name" validate:"required"`
	Description  string    `json:"description"`
	ClusterId    ClusterId `json:"clusterId" validate:"required"`
	AppGroupType string    `json:"appGroupType" validate:"oneof=LMA SERVICE_MESH"`
}

type CreateAppGroupResponse struct {
	ID string `json:"id"`
}

type CreateApplicationRequest struct {
	ApplicationType string `json:"applicationType"`
	Endpoint        string `json:"endpoint"`
	Metadata        string `json:"metadata"`
}

type GetAppGroupsResponse struct {
	AppGroups []AppGroupResponse `json:"appGroups"`
}

type GetAppGroupResponse struct {
	AppGroup AppGroupResponse `json:"appGroup"`
}

type GetApplicationsResponse struct {
	Applications []ApplicationResponse `json:"applications"`
}
