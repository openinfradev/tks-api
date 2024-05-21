package model

import (
	"encoding/json"
	"slices"
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

	ParameterSchema string  `gorm:"type:text"`
	Rego            string  `gorm:"type:text"`
	Libs            string  `gorm:"type:text"`
	SyncKinds       *string `gorm:"type:text"`
	SyncJson        *string `gorm:"type:text"`
}

type PolicyTemplate struct {
	gorm.Model

	ID                       uuid.UUID `gorm:"primarykey;type:varchar(36);not null"`
	TemplateName             string
	Type                     string                           // Org or Tks
	Version                  string                           `gorm:"-:all"` // 삭제 예정
	CurrentVersion           string                           `gorm:"-:all"`
	LatestVersion            string                           `gorm:"-:all"`
	SupportedVersions        []PolicyTemplateSupportedVersion `gorm:"foreignKey:PolicyTemplateId"`
	OrganizationId           *string                          // Org 인 경우에만 설정
	Organization             Organization                     `gorm:"foreignKey:OrganizationId"`
	Description              string
	Kind                     string
	Deprecated               bool
	Mandatory                bool // Tks 인 경우에는 무시
	Severity                 string
	PermittedOrganizations   []Organization         `gorm:"many2many:policy_template_permitted_organizations"`
	ParametersSchema         []*domain.ParameterDef `gorm:"-:all"`
	Rego                     string                 `gorm:"-:all"`
	Libs                     []string               `gorm:"-:all"`
	SyncKinds                *[]string              `gorm:"-:all"`
	SyncJson                 *string                `gorm:"-:all"`
	PermittedOrganizationIds []string               `gorm:"-:all"`
	CreatorId                *uuid.UUID             `gorm:"type:uuid"`
	Creator                  User                   `gorm:"foreignKey:CreatorId"`
	UpdatorId                *uuid.UUID             `gorm:"type:uuid"`
	Updator                  User                   `gorm:"foreignKey:UpdatorId"`
}

func (pt *PolicyTemplate) IsTksTemplate() bool {
	return strings.ToLower(pt.Type) == "tks"
}

func (pt *PolicyTemplate) IsOrganizationTemplate() bool {
	return !pt.IsTksTemplate()
}

func (pt *PolicyTemplate) ResoureName() string {
	return strings.ToLower(pt.Kind)
}

func (pt *PolicyTemplate) IsPermittedToOrganization(organizationId *string) bool {
	// tks Admin은 organizationId가 nil
	if organizationId == nil {
		return true
	}

	if pt.IsTksTemplate() {
		return slices.Contains(pt.PermittedOrganizationIds, *organizationId)
	}

	return pt.OrganizationId != nil && *organizationId == *pt.OrganizationId
}

func (pt *PolicyTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}

	jsonByte, err := json.Marshal(pt.ParametersSchema)

	if err != nil {
		return err
	}

	libs := strings.Join(pt.Libs, FILE_DELIMETER)

	pt.Version = "v1.0.0"

	var syncKindsString *string = nil

	if pt.SyncKinds != nil {
		syncJsonBytes, err := json.Marshal(pt.SyncKinds)

		if err == nil {
			syncStr := string(syncJsonBytes)
			syncKindsString = &syncStr
		}
	}

	pt.SupportedVersions = []PolicyTemplateSupportedVersion{
		{
			Version:         "v1.0.0",
			ParameterSchema: string(jsonByte),
			Rego:            pt.Rego,
			Libs:            libs,
			SyncJson:        pt.SyncJson,
			SyncKinds:       syncKindsString,
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
		pt.SyncJson = supportedVersion.SyncJson

		if len(strings.TrimSpace(supportedVersion.Libs)) == 0 {
			pt.Libs = []string{}
		} else {
			pt.Libs = strings.Split(supportedVersion.Libs, FILE_DELIMETER)
		}

		// 마찬가지로 에러 무시
		_ = json.Unmarshal([]byte(supportedVersion.ParameterSchema), &pt.ParametersSchema)

		if supportedVersion.SyncKinds != nil {
			syncKinds := []string{}
			_ = json.Unmarshal([]byte(*supportedVersion.SyncKinds), &syncKinds)
			pt.SyncKinds = &syncKinds
		}
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
