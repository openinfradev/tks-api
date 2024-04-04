package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SystemNotificationCondition struct {
	gorm.Model

	SystemNotificationRuleId uuid.UUID
	Severity                 string
	Duration                 string
	Parameter                datatypes.JSON
	Parameters               []domain.SystemNotificationParameter `gorm:"-:all"`
	EnableEmail              bool                                 `gorm:"default:false"`
	EnablePortal             bool                                 `gorm:"default:true"`
}

type SystemNotificationRule struct {
	gorm.Model

	ID                           uuid.UUID `gorm:"primarykey"`
	Name                         string    `gorm:"index,unique"`
	Description                  string
	OrganizationId               string
	Organization                 Organization `gorm:"foreignKey:OrganizationId"`
	IsSystem                     bool         `gorm:"default:false"`
	SystemNotificationTemplateId uuid.UUID
	SystemNotificationTemplate   SystemNotificationTemplate  `gorm:"foreignKey:SystemNotificationTemplateId"`
	SystemNotificationCondition  SystemNotificationCondition `gorm:"foreignKey:SystemNotificationRuleId"`
	TargetUsers                  []User                      `gorm:"many2many:system_notification_rule_users;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	TargetUserIds                []string                    `gorm:"-:all"`
	MessageTitle                 string
	MessageContent               string
	MessageActionProposal        string
	Status                       domain.SystemNotificationRuleStatus
	CreatorId                    *uuid.UUID `gorm:"type:uuid"`
	Creator                      *User      `gorm:"foreignKey:CreatorId"`
	UpdatorId                    *uuid.UUID `gorm:"type:uuid"`
	Updator                      *User      `gorm:"foreignKey:UpdatorId"`
}
