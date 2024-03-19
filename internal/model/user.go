package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId         string    `json:"accountId"`
	Password          string    `gorm:"-:all" json:"password"`
	Name              string    `json:"name"`
	Token             string    `json:"token"`
	RoleId            string
	Role              Role `gorm:"foreignKey:RoleId;references:ID" json:"role"`
	OrganizationId    string
	Organization      Organization `gorm:"foreignKey:OrganizationId;references:ID" json:"organization"`
	Creator           string       `json:"creator"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	PasswordUpdatedAt time.Time    `json:"passwordUpdatedAt"`
	PasswordExpired   bool         `json:"passwordExpired"`

	Email       string `json:"email"`
	Department  string `json:"department"`
	Description string `json:"description"`
}
