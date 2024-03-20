package domain

import (
	"github.com/google/uuid"
)

type PermissionResponse struct {
	ID        uuid.UUID             `json:"ID"`
	Name      string                `json:"name"`
	IsAllowed *bool                 `json:"is_allowed,omitempty"`
	Endpoints []*EndpointResponse   `json:"endpoints,omitempty"`
	Children  []*PermissionResponse `json:"children,omitempty"`
}

type PermissionSetResponse struct {
	Dashboard         *PermissionResponse `json:"dashboard,omitempty"`
	Stack             *PermissionResponse `json:"stack,omitempty"`
	SecurityPolicy    *PermissionResponse `json:"security_policy,omitempty"`
	ProjectManagement *PermissionResponse `json:"project_management,omitempty"`
	Notification      *PermissionResponse `json:"notification,omitempty"`
	Configuration     *PermissionResponse `json:"configuration,omitempty"`
}

type GetPermissionTemplatesResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
}

type GetPermissionsByRoleIdResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
}

type UpdatePermissionsByRoleIdRequest struct {
	Permissions []*PermissionResponse `json:"permissions"`
}
