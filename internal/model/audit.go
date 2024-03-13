package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Models
type Audit struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Group          string
	Message        string
	Description    string
	ClientIP       string
	UserId         *uuid.UUID `gorm:"type:uuid"`
	User           User       `gorm:"foreignKey:UserId"`
}

func (c *Audit) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
