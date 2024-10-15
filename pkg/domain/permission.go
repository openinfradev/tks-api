package domain

import (
	"github.com/google/uuid"
)

type GetPermissionTemplatesResponse struct {
	//Permissions *PermissionTemplateResponse `json:"permissions"`
	Permissions []*TemplateResponse `json:"permissions"`
}

type TemplateResponse struct {
	Name      string              `json:"name"`
	Key       string              `json:"key"`
	IsAllowed *bool               `json:"isAllowed,omitempty"`
	Children  []*TemplateResponse `json:"children,omitempty"`
}

type GetPermissionsByRoleIdResponse struct {
	//Permissions *PermissionSetResponse `json:"permissions"`
	Permissions []*PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID        *uuid.UUID            `json:"ID,omitempty"`
	Name      string                `json:"name"`
	Key       string                `json:"key"`
	IsAllowed *bool                 `json:"isAllowed,omitempty"`
	Endpoints []*EndpointResponse   `json:"endpoints,omitempty"`
	Children  []*PermissionResponse `json:"children,omitempty"`
}

type UpdatePermissionUpdateRequest struct {
	ID        uuid.UUID `json:"ID" validate:"required"`
	IsAllowed *bool     `json:"isAllowed" validate:"required,oneof=true false"`
}

type UpdatePermissionsByRoleIdRequest struct {
	Permissions []*UpdatePermissionUpdateRequest `json:"permissions"`
}

type GetUsersPermissionsResponse struct {
	//Permissions *MergedPermissionSetResponse `json:"permissions"`
	Permissions []*MergePermissionResponse `json:"permissions"`
}

type MergePermissionResponse struct {
	Key       string                     `json:"key"`
	IsAllowed *bool                      `json:"isAllowed,omitempty"`
	Children  []*MergePermissionResponse `json:"children,omitempty"`
}
