package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoleType string

const (
	RoleTypeDefault RoleType = "default"
	RoleTypeTks     RoleType = "tks"
	RoleTypeProject RoleType = "project"
)

type RoleResponse struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	OrganizationID string               `json:"organizationId"`
	Organization   OrganizationResponse `json:"organization"`
	Type           string               `json:"type"`
	Description    string               `json:"description"`
	Creator        uuid.UUID            `json:"creator"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

type SimpleRoleResponse = struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	Name        string `json:"name" validate:"omitempty,min=0,max=100"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type CheckRoleNameResponse struct {
	IsExist bool `json:"isExist"`
}

type AppendUsersToRoleRequest struct {
	Users []uuid.UUID `json:"users" validate:"required"`
}

type RemoveUsersFromRoleRequest struct {
	Users []uuid.UUID `json:"users" validate:"required"`
}

type GetUsersInRoleIdResponse struct {
	Users      []SimpleUserResponse `json:"users"`
	Pagination PaginationResponse   `json:"pagination"`
}
