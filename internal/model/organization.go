package model

import (
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
)

type Organization struct {
	gorm.Model

	ID               string `gorm:"primarykey;type:varchar(36);not null"`
	Name             string
	Description      string
	Phone            string
	PrimaryClusterId string
	WorkflowId       string
	Status           domain.OrganizationStatus
	StatusDesc       string
	Creator          string
}
