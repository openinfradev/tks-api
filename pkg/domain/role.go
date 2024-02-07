package domain

import (
	"github.com/google/uuid"
	"time"
)

type RoleType string

const (
	RoleTypeDefault RoleType = "default"
	RoleTypeTks     RoleType = "tks"
	RoleTypeProject RoleType = "project"
)

type Role struct {
	ID             uuid.UUID    `json:"id"`
	Name           string       `json:"name"`
	OrganizationID string       `json:"organizationId"`
	Organization   Organization `json:"organization"`
	Type           string       `json:"type"`
	Description    string       `json:"description"`
	Creator        uuid.UUID    `json:"creator"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type TksRole struct {
	RoleID uuid.UUID `json:"roleId"`
	Role
}

type ProjectRole struct {
	RoleID    uuid.UUID `json:"roleId"`
	Role      Role      `json:"role"`
	ProjectID uuid.UUID `json:"projectID"`
	Project   Project   `json:"project"`
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
