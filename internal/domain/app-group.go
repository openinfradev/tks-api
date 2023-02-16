package domain

import (
	"time"
)

// enum
type AppGroupStatus int32

const (
	AppGroupStatus_UNSPECIFIED AppGroupStatus = iota
	AppGroupStatus_INSTALLING
	AppGroupStatus_RUNNING
	AppGroupStatus_DELETING
	AppGroupStatus_DELETED
	AppGroupStatus_ERROR
)

var appGroupStatus = [...]string{
	"UNSPECIFIED",
	"INSTALLING",
	"RUNNING",
	"DELETING",
	"DELETED",
	"ERROR",
}

func (m AppGroupStatus) String() string { return appGroupStatus[(m)] }

type AppGroup = struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	ClusterId         string    `json:"clusterId"`
	AppGroupType      string    `json:"appGroupType"`
	Description       string    `json:"description"`
	WorkflowId        string    `json:"workflowId"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"created"`
	UpdatedAt         time.Time `json:"updated"`
}
