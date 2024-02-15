package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Endpoint struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;" json:"id"`
	Name         string      `gorm:"type:text;not null;unique" json:"name"`
	Group        string      `gorm:"type:text;" json:"group"`
	PermissionID uuid.UUID   `gorm:"type:uuid;" json:"permissionId"`
	Permission   *Permission `gorm:"foreignKey:PermissionID;" json:"permission"`
}

func (e *Endpoint) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.New()
	return nil
}
