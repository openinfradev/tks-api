package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SystemNotificationCondition struct {
	SystemNotificationRuleId uuid.UUID `gorm:"primarykey"`
	Order                    int       `gorm:"primarykey"`
	Severity                 string
	Duration                 int
	Condition                datatypes.JSON
	Conditions               []string `gorm:"-:all"`
	EnableEmail              bool     `gorm:"default:false"`
	EnablePortal             bool     `gorm:"default:true"`
}

type SystemNotificationRule struct {
	gorm.Model

	ID                           uuid.UUID `gorm:"primarykey"`
	Name                         string    `gorm:"index,unique"`
	Description                  string
	OrganizationId               string
	Organization                 Organization               `gorm:"foreignKey:OrganizationId"`
	SystemNotificationTemplate   SystemNotificationTemplate `gorm:"foreignKey:SystemNotificationTemplateId"`
	SystemNotificationTemplateId string
	SystemNotificationConditions []SystemNotificationCondition `gorm:"foreignKey:SystemNotificationRuleId;references:ID"`
	TargetUsers                  []User                        `gorm:"many2many:system_notification_rule_users"`
	TargetUserIds                []string                      `gorm:"-:all"`
	MessageTitle                 string
	MessageContent               string
	MessageCondition             datatypes.JSON
	MessageActionProposal        string
	CreatorId                    *uuid.UUID `gorm:"type:uuid"`
	Creator                      *User      `gorm:"foreignKey:CreatorId"`
	UpdatorId                    *uuid.UUID `gorm:"type:uuid"`
	Updator                      *User      `gorm:"foreignKey:UpdatorId"`
}
