package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dashboard struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primarykey;type:uuid"`
	OrganizationId string    `gorm:"type:varchar(36)"`
	UserId         uuid.UUID
	Content        string
	IsAdmin        bool `gorm:"default:false"`
}

func (d *Dashboard) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New()
	return nil
}
