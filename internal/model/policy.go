package model

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

type Policy struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey;type:varchar(36);not null"`
	OrganizationId string

	PolicyName  string
	Mandatory   bool
	Description string

	TargetClusterIds []string  `gorm:"-:all"`
	TargetClusters   []Cluster `gorm:"many2many:policy_target_clusters"`

	EnforcementAction string

	Parameters  string        `gorm:"type:text"`
	PolicyMatch string        `gorm:"type:text"`
	Match       *domain.Match `gorm:"-:all"`

	TemplateName   string         `gorm:"-:all"`
	TemplateId     uuid.UUID      `gorm:"type:uuid"`
	PolicyTemplate PolicyTemplate `gorm:"foreignKey:TemplateId"`

	CreatorId *uuid.UUID `gorm:"type:uuid"`
	Creator   User       `gorm:"foreignKey:CreatorId"`
	UpdatorId *uuid.UUID `gorm:"type:uuid"`
	Updator   User       `gorm:"foreignKey:UpdatorId"`
}

func (p *Policy) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()

	if p.Match != nil {
		jsonBytes, err := json.Marshal(p.Match)

		if err != nil {
			return err
		}

		p.PolicyMatch = string(jsonBytes)
	}

	return nil
}

func (p *Policy) AfterFind(tx *gorm.DB) (err error) {
	p.TemplateName = p.PolicyTemplate.TemplateName

	if len(p.PolicyMatch) > 0 {
		// 목록 조회 시 에러가 발생해서 전체 조회가 실패하는 것을 방지하기 위해서 에러는 무시
		_ = json.Unmarshal([]byte(p.PolicyMatch), p.Match)
	}

	p.TargetClusterIds = make([]string, len(p.TargetClusters))
	for i, cluster := range p.TargetClusters {
		p.TargetClusterIds[i] = cluster.ID.String()
	}

	return
}

type PolicyTargetCluster struct {
	PolicyId  uuid.UUID        `gorm:"primarykey"`
	ClusterId domain.ClusterId `gorm:"primarykey"`
}
