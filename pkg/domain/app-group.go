package domain

import (
	"time"
)

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
	"EP_UNSPECIFIED",
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

type AppGroup = struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ClusterId         string    `json:"clusterId"`
	AppGroupType      string    `json:"appGroupType"`
	Description       string    `json:"description"`
	WorkflowId        string    `json:"workflowId"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateAppGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ClusterId   string `json:"clusterId"`
	Type        string `json:"type"`
	Creator     string `json:"creator"`
}
