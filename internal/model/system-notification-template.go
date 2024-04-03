package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Models
type SystemNotificationMetricParameter struct {
	gorm.Model

	SystemNotificationTemplateId uuid.UUID
	Order                        int
	Key                          string
	Value                        string
}

type SystemNotificationTemplate struct {
	gorm.Model

	ID               uuid.UUID      `gorm:"primarykey"`
	Name             string         `gorm:"index:idx_name,unique"`
	IsSystem         bool           `gorm:"default:false"`
	Organizations    []Organization `gorm:"many2many:system_notification_template_organizations;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	OrganizationIds  []string       `gorm:"-:all"`
	Description      string
	MetricQuery      string
	MetricParameters []SystemNotificationMetricParameter `gorm:"foreignKey:SystemNotificationTemplateId;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	CreatorId        *uuid.UUID
	Creator          User `gorm:"foreignKey:CreatorId"`
	UpdatorId        *uuid.UUID
	Updator          User `gorm:"foreignKey:UpdatorId"`
}
