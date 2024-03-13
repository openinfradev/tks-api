package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

type PolicyTemplateSupportedVersion struct {
	gorm.Model

	PolicyTemplateId uuid.UUID `gorm:"index:template_version,unique"`
	Version          string    `gorm:"index:template_version,unique"`

	ParameterSchema string `gorm:"type:text"`
	Rego            string `gorm:"type:text"`
	Libs            string `gorm:"type:text"`
}

type PolicyTemplate struct {
	gorm.Model

	ID                       uuid.UUID `gorm:"primarykey;type:varchar(36);not null"`
	Type                     string    // Org or Tks
	Name                     string
	Version                  string
	SupportedVersions        []PolicyTemplateSupportedVersion `gorm:"foreignKey:PolicyTemplateId"`
	Description              string
	Kind                     string
	Deprecated               bool
	Mandatory                bool // Tks 인 경우에는 무시
	Severity                 string
	PermittedOrganizations   []Organization        `gorm:"many2many:policy_template_permitted_organiations;"`
	TemplateName             string                `gorm:"-:all" json:"templateName"`
	ParametersSchema         []domain.ParameterDef `gorm:"-:all" json:"parametersSchema,omitempty"`
	Rego                     string                `gorm:"-:all"`
	Libs                     []string              `gorm:"-:all"`
	PermittedOrganizationIds []string              `gorm:"-:all"`
	CreatorId                *uuid.UUID            `gorm:"type:uuid"`
	Creator                  User                  `gorm:"foreignKey:CreatorId"`
	UpdatorId                *uuid.UUID            `gorm:"type:uuid"`
	Updator                  User                  `gorm:"foreignKey:UpdatorId"`
}

func (c *PolicyTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
