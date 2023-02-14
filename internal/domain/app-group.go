package domain

import (
	"time"
)

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
