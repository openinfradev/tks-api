package repository

import "gorm.io/gorm"

// Models
type Workflow struct {
	gorm.Model

	RefID      string `gorm:"uniqueIndex"`
	RefType    string // cluster, organization, appgroup
	WorkflowId string
	StatusDesc string
}
