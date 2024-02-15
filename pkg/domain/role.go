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

type TksRole struct {
	RoleID string `gorm:"primarykey;" json:"roleId"`
	Role   Role   `gorm:"foreignKey:RoleID;references:ID;"`
}

type ProjectRole struct {
	RoleID    string  `gorm:"primaryKey" json:"roleId"`
	Role      Role    `gorm:"foreignKey:RoleID;references:ID;" json:"role"`
	ProjectID string  `json:"projectID"`
	Project   Project `gorm:"foreignKey:ProjectID;references:ID;" json:"project"`
}

//type Role = struct {
//	ID             uuid.UUID    `json:"id"`
//	Name           string       `json:"name"`
//	OrganizationID string       `json:"organizationId"`
//	Organization   Organization `json:"organization"`
//	Type           string       `json:"type"`
//	Description    string       `json:"description"`
//	Creator        uuid.UUID    `json:"creator"`
//	CreatedAt      time.Time    `json:"createdAt"`
//	UpdatedAt      time.Time    `json:"updatedAt"`
//}

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

type CreateProjectRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type CreateProjectRoleResponse struct {
	ID string `json:"id"`
}

type GetProjectRoleResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Description    string    `json:"description"`
	Creator        string    `json:"creator"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ListProjectRoleResponse struct {
	Roles      []GetProjectRoleResponse `json:"roles"`
	Pagination PaginationResponse       `json:"pagination"`
}

type UpdateProjectRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}
