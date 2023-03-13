package repository

import "gorm.io/gorm"

// Models
type Workflow struct {
	gorm.Model

	RefID      string
	RefType    string // cluster, organization, appgroup
	WorkflowId string
	StatusDesc string
}
