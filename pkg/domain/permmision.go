package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model

	ID   uuid.UUID `gorm:"primarykey;type:uuid;" json:"id"`
	Name string    `json:"name"`

	IsAllowed *bool       `gorm:"type:boolean;" json:"is_allowed,omitempty"`
	RoleID    *string     `json:"role_id,omitempty"`
	Role      *Role       `gorm:"foreignKey:RoleID;references:ID;" json:"role,omitempty"`
	Endpoints []*Endpoint `gorm:"one2many:endpoints;" json:"endpoints,omitempty"`
	// omit empty

	ParentID *uuid.UUID    `json:"parent_id,omitempty"`
	Parent   *Permission   `gorm:"foreignKey:ParentID;references:ID;" json:"parent,omitempty"`
	Children []*Permission `gorm:"foreignKey:ParentID;references:ID;" json:"children,omitempty"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
