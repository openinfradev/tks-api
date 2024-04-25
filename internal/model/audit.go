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
	Organization   Organization
	Group          string
	Message        string
	Description    string
	ClientIP       string
	UserId         *uuid.UUID `gorm:"type:uuid"`
	UserAccountId  string
	UserName       string
	UserRoles      string
}
