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
	EnableEmail              bool `gorm:"default:false"`
	EnablePortal             bool `gorm:"default:true"`
}

type SystemNotificationMessage struct {
	SystemNotificationRuleId uuid.UUID `gorm:"primarykey"`
	Title                    string
	Content                  int
	Condition                datatypes.JSON
	ActionProposal           string
	TargetUsers              []User `gorm:"many2many:system_notification_message_users"`
}

type SystemNotificationRule struct {
	gorm.Model

	ID          uuid.UUID `gorm:"primarykey"`
	Name        string    `gorm:"index,unique"`
	Description string
	Templates   []SystemNotificationTemplate `gorm:"many2many:system_notification_rule_system_notification_templates"`
	Messages    []SystemNotificationMessage  `gorm:"many2many:system_notification_rule_system_notification_messages"`

	CreatorId *uuid.UUID `gorm:"type:uuid"`
	Creator   *User      `gorm:"foreignKey:CreatorId"`
	UpdatorId *uuid.UUID `gorm:"type:uuid"`
	Updator   *User      `gorm:"foreignKey:UpdatorId"`
}

func (c *SystemNotificationRule) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
