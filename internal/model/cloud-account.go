package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

// Models
type CloudAccount struct {
	gorm.Model

	ID              uuid.UUID `gorm:"primarykey"`
	OrganizationId  string
	Organization    Organization `gorm:"foreignKey:OrganizationId"`
	Name            string       `gorm:"index"`
	Description     string       `gorm:"index"`
	Resource        string
	CloudService    string
	WorkflowId      string
	Status          domain.CloudAccountStatus
	StatusDesc      string
	AwsAccountId    string
	AccessKeyId     string `gorm:"-:all"`
	SecretAccessKey string `gorm:"-:all"`
	SessionToken    string `gorm:"-:all"`
	Clusters        int    `gorm:"-:all"`
	CreatedIAM      bool
	CreatorId       *uuid.UUID `gorm:"type:uuid"`
	Creator         User       `gorm:"foreignKey:CreatorId"`
	UpdatorId       *uuid.UUID `gorm:"type:uuid"`
	Updator         User       `gorm:"foreignKey:UpdatorId"`
}
