package model

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type Organization struct {
	gorm.Model

	ID                          string `gorm:"primarykey;type:varchar(36);not null"`
	Name                        string
	Description                 string
	Phone                       string
	PrimaryClusterId            string
	WorkflowId                  string
	Status                      domain.OrganizationStatus
	StatusDesc                  string
	CreatorId                   *uuid.UUID                   `gorm:"type:uuid"`
	StackTemplates              []StackTemplate              `gorm:"many2many:stack_template_organizations"`
	SystemNotificationTemplates []SystemNotificationTemplate `gorm:"many2many:system_notification_template_organizations"`
	PolicyTemplates             []PolicyTemplate             `gorm:"many2many:policy_template_permitted_organiations;"`
	ClusterCount                int                          `gorm:"-:all"`
	AdminId                     *uuid.UUID
	Admin                       *User `gorm:"foreignKey:AdminId"`
}
