package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	ID             string       `gorm:"primarykey;" json:"id"`
	Name           string       `json:"name"`
	OrganizationID string       `json:"organizationId"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID;" json:"organization"`
	Type           string       `json:"type"`
	Description    string       `json:"description"`
	Creator        uuid.UUID    `json:"creator"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}
