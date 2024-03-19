package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StackTemplate struct {
	gorm.Model

	ID              uuid.UUID `gorm:"primarykey"`
	Name            string    `gorm:"index"`
	Description     string    `gorm:"index"`
	Template        string
	TemplateType    string
	Version         string
	CloudService    string
	Platform        string
	KubeVersion     string
	KubeType        string
	Organizations   []Organization `gorm:"many2many:stack_template_organizations"`
	Services        datatypes.JSON
	ServiceIds      []string   `gorm:"-:all"`
	OrganizationIds []string   `gorm:"-:all"`
	CreatorId       *uuid.UUID `gorm:"type:uuid"`
	Creator         User       `gorm:"foreignKey:CreatorId"`
	UpdatorId       *uuid.UUID `gorm:"type:uuid"`
	Updator         User       `gorm:"foreignKey:UpdatorId"`
}

type StackTemplateOrganization struct {
	StackTemplateId uuid.UUID `gorm:"primarykey"`
	OrganizationId  string    `gorm:"primarykey"`
}
