package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type RoleType string

const (
	RoleTypeDefault RoleType = "default"
	RoleTypeTks     RoleType = "tks"
	RoleTypeProject RoleType = "project"
)

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New().String()
	return nil
}

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

type CreateTksRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type CreateTksRoleResponse struct {
	ID string `json:"id"`
}

type GetTksRoleResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	Description    string    `json:"description"`
	Creator        string    `json:"creator"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ListTksRoleResponse struct {
	Roles      []GetTksRoleResponse `json:"roles"`
	Pagination PaginationResponse   `json:"pagination"`
}

type UpdateTksRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}
