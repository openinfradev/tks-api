package http

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IPermissionHandler interface {
	GetPermissionTemplates(w http.ResponseWriter, r *http.Request)
	GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request)
	UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request)
}

type PermissionHandler struct {
	permissionUsecase usecase.IPermissionUsecase
	userUsecase       usecase.IUserUsecase
}

func NewPermissionHandler(usecase usecase.Usecase) *PermissionHandler {
	return &PermissionHandler{
		permissionUsecase: usecase.Permission,
		userUsecase:       usecase.User,
	}
}

// GetPermissionTemplates godoc
//
//	@Tags			Permission
//	@Summary		Get Permission Templates
//	@Description	Get Permission Templates
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.PermissionSetResponse
//	@Router			/permissions/templates [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionTemplates(w http.ResponseWriter, r *http.Request) {
	permissionSet := model.NewDefaultPermissionSet()

	var out domain.GetPermissionTemplatesResponse
	out.Permissions = new(domain.PermissionTemplateResponse)

	out.Permissions.Dashboard = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Dashboard)
	out.Permissions.Stack = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Stack)
	out.Permissions.Policy = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Policy)
	out.Permissions.ProjectManagement = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.ProjectManagement)
	out.Permissions.Notification = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Notification)
	out.Permissions.Configuration = convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Configuration)

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertModelToPermissionTemplateResponse(ctx context.Context, permission *model.Permission) *domain.TemplateResponse {
	var permissionResponse domain.TemplateResponse

	permissionResponse.Key = permission.Key
	permissionResponse.Name = permission.Name
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToPermissionTemplateResponse(ctx, child))
	}

	return &permissionResponse
}

// GetPermissionsByAccountId godoc
//
//	@Tags			Permission
//	@Summary		Get Permissions By Account ID
//	@Description	Get Permissions By Account ID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.GetUsersPermissionsResponse
//	@Router			/organizations/{organizationId}/users/{accountId}/permissions [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionsByAccountId(w http.ResponseWriter, r *http.Request) {
	var organizationId, accountId string

	vars := mux.Vars(r)
	if v, ok := vars["accountId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		accountId = v
	}
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	user, err := h.userUsecase.GetByAccountId(r.Context(), accountId, organizationId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	var roles []*model.Role
	roles = append(roles, &user.Role)

	var permissionSets []*model.PermissionSet
	for _, role := range roles {
		permissionSet, err := h.permissionUsecase.GetPermissionSetByRoleId(r.Context(), role.ID)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		permissionSets = append(permissionSets, permissionSet)
	}

	mergedPermissionSet := h.permissionUsecase.MergePermissionWithOrOperator(r.Context(), permissionSets...)

	var permissions domain.MergedPermissionSetResponse
	permissions.Dashboard = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Dashboard)
	permissions.Stack = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Stack)
	permissions.Policy = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Policy)
	permissions.ProjectManagement = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.ProjectManagement)
	permissions.Notification = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Notification)
	permissions.Configuration = convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Configuration)

	var out domain.GetUsersPermissionsResponse
	out.Permissions = &permissions
	ResponseJSON(w, r, http.StatusOK, out)

}

func convertModelToMergedPermissionSetResponse(ctx context.Context, permission *model.Permission) *domain.MergePermissionResponse {
	var permissionResponse domain.MergePermissionResponse

	permissionResponse.Key = permission.Key
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToMergedPermissionSetResponse(ctx, child))
	}

	return &permissionResponse
}

// GetPermissionsByRoleId godoc
//
//	@Tags			Permission
//	@Summary		Get Permissions By Role ID
//	@Description	Get Permissions By Role ID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.PermissionSetResponse
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var roleId string

	vars := mux.Vars(r)
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	permissionSet, err := h.permissionUsecase.GetPermissionSetByRoleId(r.Context(), roleId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	var permissionSetResponse domain.PermissionSetResponse
	permissionSetResponse.Dashboard = convertModelToPermissionResponse(r.Context(), permissionSet.Dashboard)
	permissionSetResponse.Stack = convertModelToPermissionResponse(r.Context(), permissionSet.Stack)
	permissionSetResponse.Policy = convertModelToPermissionResponse(r.Context(), permissionSet.Policy)
	permissionSetResponse.ProjectManagement = convertModelToPermissionResponse(r.Context(), permissionSet.ProjectManagement)
	permissionSetResponse.Notification = convertModelToPermissionResponse(r.Context(), permissionSet.Notification)
	permissionSetResponse.Configuration = convertModelToPermissionResponse(r.Context(), permissionSet.Configuration)

	var out domain.GetPermissionsByRoleIdResponse
	out.Permissions = &permissionSetResponse

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertModelToPermissionResponse(ctx context.Context, permission *model.Permission) *domain.PermissionResponse {
	var permissionResponse domain.PermissionResponse

	permissionResponse.ID = permission.ID
	permissionResponse.Key = permission.Key
	permissionResponse.Name = permission.Name
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
	}

	for _, endpoint := range permission.Endpoints {
		permissionResponse.Endpoints = append(permissionResponse.Endpoints, convertModelToEndpointResponse(ctx, endpoint))
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToPermissionResponse(ctx, child))
	}

	return &permissionResponse
}

func convertModelToEndpointResponse(ctx context.Context, endpoint *model.Endpoint) *domain.EndpointResponse {
	var endpointResponse domain.EndpointResponse

	endpointResponse.Name = endpoint.Name
	endpointResponse.Group = endpoint.Group

	return &endpointResponse
}

// UpdatePermissionsByRoleId godoc
//
//	@Tags			Permission
//	@Summary		Update Permissions By Role ID
//	@Description	Update Permissions By Role ID
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path	string									true	"Role ID"
//	@Param			body	body	domain.UpdatePermissionsByRoleIdRequest	true	"Update Permissions By Role ID Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [put]
//	@Security		JWT
func (h PermissionHandler) UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
	// request
	input := domain.UpdatePermissionsByRoleIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	for _, permissionResponse := range input.Permissions {
		var permission model.Permission
		permission.ID = permissionResponse.ID
		permission.IsAllowed = permissionResponse.IsAllowed

		if err := h.permissionUsecase.UpdatePermission(r.Context(), &permission); err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
