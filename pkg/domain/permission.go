package domain

import (
	"github.com/google/uuid"
)

type GetPermissionTemplatesResponse struct {
	//Permissions *PermissionTemplateResponse `json:"permissions"`
	Permissions []*TemplateResponse `json:"permissions"`
}

//type PermissionTemplateResponse struct {
//	Dashboard         *TemplateResponse `json:"dashboard,omitempty"`
//	Stack             *TemplateResponse `json:"stack,omitempty"`
//	Policy            *TemplateResponse `json:"policy,omitempty"`
//	ProjectManagement *TemplateResponse `json:"project_management,omitempty"`
//	Notification      *TemplateResponse `json:"notification,omitempty"`
//	Configuration     *TemplateResponse `json:"configuration,omitempty"`
//}

type TemplateResponse struct {
	Name     string              `json:"name"`
	Key      string              `json:"key"`
	EdgeKey  *string             `json:"edgeKey,omitempty"`
	Children []*TemplateResponse `json:"children,omitempty"`
}

type GetPermissionsByRoleIdResponse struct {
	//Permissions *PermissionSetResponse `json:"permissions"`
	Permissions []*PermissionResponse `json:"permissions"`
}

//type PermissionSetResponse struct {
//	Dashboard         *PermissionResponse   `json:"dashboard,omitempty"`
//	Stack             *PermissionResponse   `json:"stack,omitempty"`
//	Policy            *PermissionResponse   `json:"policy,omitempty"`
//	ProjectManagement *PermissionResponse   `json:"project_management,omitempty"`
//	Notification      *PermissionResponse   `json:"notification,omitempty"`
//	Configuration     *PermissionResponse   `json:"configuration,omitempty"`
//}

type PermissionResponse struct {
	ID        *uuid.UUID            `json:"ID,omitempty"`
	Name      string                `json:"name"`
	Key       string                `json:"key"`
	IsAllowed *bool                 `json:"isAllowed,omitempty"`
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

//type MergedPermissionSetResponse struct {
//	Dashboard         *MergePermissionResponse `json:"dashboard,omitempty"`
//	Stack             *MergePermissionResponse `json:"stack,omitempty"`
//	Policy            *MergePermissionResponse `json:"policy,omitempty"`
//	ProjectManagement *MergePermissionResponse `json:"project_management,omitempty"`
//	Notification      *MergePermissionResponse `json:"notification,omitempty"`
//	Configuration     *MergePermissionResponse `json:"configuration,omitempty"`
//}

type MergePermissionResponse struct {
	Key       string                     `json:"key"`
	IsAllowed *bool                      `json:"isAllowed,omitempty"`
	Children  []*MergePermissionResponse `json:"children,omitempty"`
}

type GetPermissionEdgeKeysResponse struct {
}

type GetEndpointsResponse struct {
	Endpoints []EndpointResponse `json:"endpoints"`
}
