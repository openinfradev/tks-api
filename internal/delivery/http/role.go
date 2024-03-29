package http

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IRoleHandler interface {
	CreateTksRole(w http.ResponseWriter, r *http.Request)
	ListTksRoles(w http.ResponseWriter, r *http.Request)
	GetTksRole(w http.ResponseWriter, r *http.Request)
	DeleteTksRole(w http.ResponseWriter, r *http.Request)
	UpdateTksRole(w http.ResponseWriter, r *http.Request)

	GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request)
	UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request)

	Admin_ListTksRoles(w http.ResponseWriter, r *http.Request)
	Admin_GetTksRole(w http.ResponseWriter, r *http.Request)
}

type RoleHandler struct {
	roleUsecase       usecase.IRoleUsecase
	permissionUsecase usecase.IPermissionUsecase
}

func NewRoleHandler(usecase usecase.Usecase) *RoleHandler {
	return &RoleHandler{
		roleUsecase:       usecase.Role,
		permissionUsecase: usecase.Permission,
	}
}

// CreateTksRole godoc
//
//	@Tags			Roles
//	@Summary		Create Tks Role
//	@Description	Create Tks Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"Organization ID"
//	@Param			body			body		domain.CreateTksRoleRequest	true	"Create Tks Role Request"
//	@Success		200				{object}	domain.CreateTksRoleResponse
//	@Router			/organizations/{organizationId}/roles [post]
//	@Security		JWT
func (h RoleHandler) CreateTksRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var organizationId string

	vars := mux.Vars(r)
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	// request body
	input := domain.CreateTksRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// input to dto
	dto := model.Role{
		OrganizationID: organizationId,
		Name:           input.Name,
		Description:    input.Description,
		Type:           string(domain.RoleTypeTks),
	}

	// create role
	var roleId string
	if roleId, err = h.roleUsecase.CreateTksRole(r.Context(), &dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// create permission
	defaultPermissionSet := model.NewDefaultPermissionSet()
	h.permissionUsecase.SetRoleIdToPermissionSet(r.Context(), roleId, defaultPermissionSet)
	err = h.permissionUsecase.CreatePermissionSet(r.Context(), defaultPermissionSet)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, domain.CreateTksRoleResponse{ID: roleId})
}

// ListTksRoles godoc
//
//	@Tags			Roles
//	@Summary		List Tks Roles
//	@Description	List Tks Roles
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.ListTksRoleResponse
//	@Router			/organizations/{organizationId}/roles [get]
//	@Security		JWT
func (h RoleHandler) ListTksRoles(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var organizationId string

	vars := mux.Vars(r)
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	// query parameter
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	// list roles
	roles, err := h.roleUsecase.ListTksRoles(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListTksRoleResponse
	out.Roles = make([]domain.GetTksRoleResponse, len(roles))
	for i, role := range roles {
		out.Roles[i] = domain.GetTksRoleResponse{
			ID:             role.ID,
			Name:           role.Name,
			OrganizationID: role.OrganizationID,
			Description:    role.Description,
			Creator:        role.Creator.String(),
			CreatedAt:      role.CreatedAt,
			UpdatedAt:      role.UpdatedAt,
		}
	}

	if err := serializer.Map(r.Context(), *pg, &out.Pagination); err != nil {
		log.Info(r.Context(), err)
	}

	// response
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetTksRole godoc
//
//	@Tags			Roles
//	@Summary		Get Tks Role
//	@Description	Get Tks Role
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			roleId			path		string	true	"Role ID"
//	@Success		200				{object}	domain.GetTksRoleResponse
//	@Router			/organizations/{organizationId}/roles/{roleId} [get]
//	@Security		JWT
func (h RoleHandler) GetTksRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		roleId = v
	}

	// get role
	role, err := h.roleUsecase.GetTksRole(r.Context(), roleId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	out := domain.GetTksRoleResponse{
		ID:             role.ID,
		Name:           role.Name,
		OrganizationID: role.OrganizationID,
		Description:    role.Description,
		Creator:        role.Creator.String(),
		CreatedAt:      role.CreatedAt,
		UpdatedAt:      role.UpdatedAt,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// DeleteTksRole godoc
//
//	@Tags			Roles
//	@Summary		Delete Tks Role
//	@Description	Delete Tks Role
//	@Produce		json
//	@Param			organizationId	path	string	true	"Organization ID"
//	@Param			roleId			path	string	true	"Role ID"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId} [delete]
//	@Security		JWT
func (h RoleHandler) DeleteTksRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// delete role
	if err := h.roleUsecase.DeleteTksRole(r.Context(), roleId); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// UpdateTksRole godoc
//
//	@Tags			Roles
//	@Summary		Update Tks Role
//	@Description	Update Tks Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string						true	"Organization ID"
//	@Param			roleId			path	string						true	"Role ID"
//	@Param			body			body	domain.UpdateTksRoleRequest	true	"Update Tks Role Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId} [put]
//	@Security		JWT
func (h RoleHandler) UpdateTksRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// request body
	input := domain.UpdateTksRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// input to dto
	dto := model.Role{
		ID:          roleId,
		Description: input.Description,
	}

	// update role
	if err := h.roleUsecase.UpdateTksRole(r.Context(), &dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetPermissionsByRoleId godoc
//
//	@Tags			Roles
//	@Summary		Get Permissions By Role ID
//	@Description	Get Permissions By Role ID
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			roleId			path		string	true	"Role ID"
//	@Success		200				{object}	domain.GetPermissionsByRoleIdResponse
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [get]
//	@Security		JWT
func (h RoleHandler) GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
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

	var out domain.GetPermissionsByRoleIdResponse
	out.Permissions = make([]*domain.PermissionResponse, 0)

	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.Dashboard))
	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.Stack))
	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.Policy))
	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.ProjectManagement))
	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.Notification))
	out.Permissions = append(out.Permissions, convertModelToPermissionResponse(r.Context(), permissionSet.Configuration))

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertModelToPermissionResponse(ctx context.Context, permission *model.Permission) *domain.PermissionResponse {
	var permissionResponse domain.PermissionResponse

	permissionResponse.Key = permission.Key
	permissionResponse.Name = permission.Name
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
		permissionResponse.ID = &permission.ID
	}

	for _, endpoint := range permission.Endpoints {
		permissionResponse.Endpoints = append(permissionResponse.Endpoints, convertModelToEndpointResponse(ctx, endpoint))
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToPermissionResponse(ctx, child))
	}

	return &permissionResponse
}

func convertModelToEndpointResponse(_ context.Context, endpoint *model.Endpoint) *domain.EndpointResponse {
	var endpointResponse domain.EndpointResponse

	endpointResponse.Name = endpoint.Name
	endpointResponse.Group = endpoint.Group

	return &endpointResponse
}

// UpdatePermissionsByRoleId godoc
//
//	@Tags			Roles
//	@Summary		Update Permissions By Role ID
//	@Description	Update Permissions By Role ID
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string									true	"Organization ID"
//	@Param			roleId			path	string									true	"Role ID"
//	@Param			body			body	domain.UpdatePermissionsByRoleIdRequest	true	"Update Permissions By Role ID Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [put]
//	@Security		JWT
func (h RoleHandler) UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
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

// Admin_ListTksRoles godoc
//
//	@Tags			Roles
//	@Summary		Admin List Tks Roles
//	@Description	Admin List Tks Roles
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.ListTksRoleResponse
//	@Router			/admin/organizations/{organizationId}/roles [get]
func (h RoleHandler) Admin_ListTksRoles(w http.ResponseWriter, r *http.Request) {
	// Same as ListTksRoles

	var organizationId string

	vars := mux.Vars(r)
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	// query parameter
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	// list roles
	roles, err := h.roleUsecase.ListTksRoles(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListTksRoleResponse
	out.Roles = make([]domain.GetTksRoleResponse, len(roles))
	for i, role := range roles {
		out.Roles[i] = domain.GetTksRoleResponse{
			ID:             role.ID,
			Name:           role.Name,
			OrganizationID: role.OrganizationID,
			Description:    role.Description,
			Creator:        role.Creator.String(),
			CreatedAt:      role.CreatedAt,
			UpdatedAt:      role.UpdatedAt,
		}
	}

	if err := serializer.Map(r.Context(), *pg, &out.Pagination); err != nil {
		log.Info(r.Context(), err)
	}

	// response
	ResponseJSON(w, r, http.StatusOK, out)
}

// Admin_GetTksRole godoc
//
//	@Tags			Roles
//	@Summary		Admin Get Tks Role
//	@Description	Admin Get Tks Role
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			roleId			path		string	true	"Role ID"
//	@Success		200				{object}	domain.GetTksRoleResponse
//	@Router			/admin/organizations/{organizationId}/roles/{roleId} [get]
func (h RoleHandler) Admin_GetTksRole(w http.ResponseWriter, r *http.Request) {
	// Same as GetTksRole

	vars := mux.Vars(r)
	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		roleId = v
	}

	// get role
	role, err := h.roleUsecase.GetTksRole(r.Context(), roleId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	out := domain.GetTksRoleResponse{
		ID:             role.ID,
		Name:           role.Name,
		OrganizationID: role.OrganizationID,
		Description:    role.Description,
		Creator:        role.Creator.String(),
		CreatedAt:      role.CreatedAt,
		UpdatedAt:      role.UpdatedAt,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
