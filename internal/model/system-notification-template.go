package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Models
type MetricParameter struct {
	SystemNotificationTemplateId uuid.UUID `gorm:"primarykey"`
	Order                        int       `gorm:"primarykey"`
	Key                          string
	Value                        string
}

type SystemNotificationTemplate struct {
	gorm.Model

	ID               uuid.UUID      `gorm:"primarykey"`
	Name             string         `gorm:"index:idx_name,unique"`
	Organizations    []Organization `gorm:"many2many:alert_template_organizations"`
	OrganizationIds  []string       `gorm:"-:all"`
	Description      string
	MetricQuery      string
	MetricParameters []MetricParameter `gorm:"foreignKey:SystemNotificationTemplateId;references:ID"`
	CreatorId        *uuid.UUID
	Creator          User `gorm:"foreignKey:CreatorId"`
	UpdatorId        *uuid.UUID
	Updator          User `gorm:"foreignKey:UpdatorId"`
}

func (c *SystemNotificationTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
