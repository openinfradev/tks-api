package model

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

const (
	FILE_DELIMETER = "---\n"
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
	TemplateName             string
	Type                     string                           // Org or Tks
	Version                  string                           `gorm:"-:all"` // 삭제 예정
	SupportedVersions        []PolicyTemplateSupportedVersion `gorm:"foreignKey:PolicyTemplateId"`
	Description              string
	Kind                     string
	Deprecated               bool
	Mandatory                bool // Tks 인 경우에는 무시
	Severity                 string
	PermittedOrganizations   []Organization        `gorm:"many2many:policy_template_permitted_organiations"`
	ParametersSchema         []domain.ParameterDef `gorm:"-:all"`
	Rego                     string                `gorm:"-:all"`
	Libs                     []string              `gorm:"-:all"`
	PermittedOrganizationIds []string              `gorm:"-:all"`
	CreatorId                *uuid.UUID            `gorm:"type:uuid"`
	Creator                  User                  `gorm:"foreignKey:CreatorId"`
	UpdatorId                *uuid.UUID            `gorm:"type:uuid"`
	Updator                  User                  `gorm:"foreignKey:UpdatorId"`
}

func (pt *PolicyTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	pt.ID = uuid.New()

	jsonByte, err := json.Marshal(pt.ParametersSchema)

	if err != nil {
		return err
	}

	libs := strings.Join(pt.Libs, FILE_DELIMETER)

	pt.Version = "v1.0.0"

	pt.SupportedVersions = []PolicyTemplateSupportedVersion{
		{
			Version:         "v1.0.0",
			ParameterSchema: string(jsonByte),
			Rego:            pt.Rego,
			Libs:            libs,
		},
	}

	return nil
}

func (pt *PolicyTemplate) AfterFind(tx *gorm.DB) (err error) {
	// 목록 조회 시 에러가 발생해서 전체 조회가 실패하는 것을 방지하기 위해서 에러는 무시
	if len(pt.SupportedVersions) > 0 {
		supportedVersion := pt.SupportedVersions[0]
		pt.Version = supportedVersion.Version
		pt.Rego = supportedVersion.Rego
		pt.Libs = strings.Split(supportedVersion.ParameterSchema, FILE_DELIMETER)

		// 마찬가지로 에러 무시
		json.Unmarshal([]byte(supportedVersion.ParameterSchema), &pt.ParametersSchema)
	}

	pt.PermittedOrganizationIds = make([]string, len(pt.PermittedOrganizations))
	for i, org := range pt.PermittedOrganizations {
		pt.PermittedOrganizationIds[i] = org.ID
	}

	return
}

type PolicyTemplatePermittedOrganization struct {
	PolicyTemplateId uuid.UUID `gorm:"primarykey"`
	OrganizationId   string    `gorm:"primarykey"`
}
