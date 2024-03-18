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
	OrganizationId            string
	Organization              Organization `gorm:"foreignKey:OrganizationId"`
	Name                      string
	Code                      string
	Description               string
	Grade                     string
	Message                   string
	ClusterId                 domain.ClusterId
	Cluster                   Cluster `gorm:"foreignKey:ClusterId"`
	Node                      string
	CheckPoint                string
	GrafanaUrl                string
	FiredAt                   *time.Time `gorm:"-:all"`
	TakedAt                   *time.Time `gorm:"-:all"`
	ClosedAt                  *time.Time `gorm:"-:all"`
	TakedSec                  int        `gorm:"-:all"`
	ProcessingSec             int        `gorm:"-:all"`
	LastTaker                 User       `gorm:"-:all"`
	SystemNotificationActions []SystemNotificationAction
	Summary                   string
	RawData                   datatypes.JSON
	Status                    domain.SystemNotificationActionStatus `gorm:"index"`
}

func (c *SystemNotification) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
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

func (c *SystemNotificationAction) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
