package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Models
type SystemNotification struct {
	gorm.Model

	ID                        uuid.UUID `gorm:"primarykey"`
	Name                      string
	NotificationType          string `gorm:"default:SYSTEM_NOTIFICATION"`
	OrganizationId            string
	Organization              Organization `gorm:"foreignKey:OrganizationId"`
	ClusterId                 domain.ClusterId
	Cluster                   Cluster `gorm:"foreignKey:ClusterId"`
	Severity                  string
	MessageTitle              string
	MessageContent            string
	MessageActionProposal     string
	Node                      string
	GrafanaUrl                string
	FiredAt                   *time.Time                 `gorm:"-:all"`
	TakedAt                   *time.Time                 `gorm:"-:all"`
	ClosedAt                  *time.Time                 `gorm:"-:all"`
	TakedSec                  int                        `gorm:"-:all"`
	ProcessingSec             int                        `gorm:"-:all"`
	LastTaker                 User                       `gorm:"-:all"`
	SystemNotificationActions []SystemNotificationAction `gorm:"foreignKey:SystemNotificationId;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	Summary                   string
	RawData                   datatypes.JSON
	Status                    domain.SystemNotificationActionStatus `gorm:"index"`
	Read                      bool                                  `gorm:"-:all"`
	Readers                   []User                                `gorm:"many2many:system_notification_users;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

type SystemNotificationAction struct {
	gorm.Model

	ID                   uuid.UUID `gorm:"primarykey"`
	SystemNotificationId uuid.UUID
	Content              string
	Status               domain.SystemNotificationActionStatus
	TakerId              *uuid.UUID `gorm:"type:uuid"`
	Taker                User       `gorm:"foreignKey:TakerId"`
}
