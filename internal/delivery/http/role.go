package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
)

type IRoleHandler interface {
	CreateTksRole(w http.ResponseWriter, r *http.Request)
	ListTksRoles(w http.ResponseWriter, r *http.Request)
	GetTksRole(w http.ResponseWriter, r *http.Request)
	DeleteTksRole(w http.ResponseWriter, r *http.Request)
	UpdateTksRole(w http.ResponseWriter, r *http.Request)
	GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request)
	UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request)
	IsRoleNameExisted(w http.ResponseWriter, r *http.Request)

	GetUsersInRoleId(w http.ResponseWriter, r *http.Request)
	AppendUsersToRole(w http.ResponseWriter, r *http.Request)
	RemoveUsersFromRole(w http.ResponseWriter, r *http.Request)

	Admin_ListTksRoles(w http.ResponseWriter, r *http.Request)
	Admin_GetTksRole(w http.ResponseWriter, r *http.Request)
}

type RoleHandler struct {
	roleUsecase       usecase.IRoleUsecase
	userUsecase       usecase.IUserUsecase
	permissionUsecase usecase.IPermissionUsecase
	stackUsecease     usecase.IStackUsecase
}

func NewRoleHandler(usecase usecase.Usecase) *RoleHandler {
	return &RoleHandler{
		roleUsecase:       usecase.Role,
		permissionUsecase: usecase.Permission,
		userUsecase:       usecase.User,
		stackUsecease:     usecase.Stack,
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
	var organizationId, roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		roleId = v
	}
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		organizationId = v
	}

	// get role
	role, err := h.roleUsecase.GetTksRole(r.Context(), organizationId, roleId)
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

	var organizationId string
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	affectedUsers, err := h.userUsecase.ListUsersByRole(r.Context(), organizationId, roleId, nil)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	// delete role
	if err := h.roleUsecase.DeleteTksRole(r.Context(), organizationId, roleId); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// Sync ClusterAdmin Permission to Keycloak
	for _, user := range *affectedUsers {
		stacks, err := h.stackUsecease.Fetch(r.Context(), organizationId, nil)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		stackIds := make([]string, 0)
		for _, stack := range stacks {
			stackIds = append(stackIds, stack.ID.String())
		}
		err = h.syncKeycloakWithClusterAdminPermission(r.Context(), organizationId, stackIds, []model.User{user})
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
		Name:        input.Name,
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
	// path parameter
	vars := mux.Vars(r)

	var organizationId string
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// request
	input := domain.UpdatePermissionsByRoleIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	var clusterAdminPermissionChanged bool

	for _, permissionResponse := range input.Permissions {
		var permission model.Permission
		permission.ID = permissionResponse.ID
		permission.IsAllowed = permissionResponse.IsAllowed

		if err := h.permissionUsecase.UpdatePermission(r.Context(), &permission); err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		if permission.Parent != nil && permission.Parent.Key == model.MiddleClusterAccessControlKey {
			clusterAdminPermissionChanged = true
		}
	}

	// Sync ClusterAdmin Permission to Keycloak
	if clusterAdminPermissionChanged {
		users, err := h.userUsecase.ListUsersByRole(r.Context(), organizationId, roleId, nil)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		for _, user := range *users {
			stacks, err := h.stackUsecease.Fetch(r.Context(), organizationId, nil)
			if err != nil {
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
			stackIds := make([]string, 0)
			for _, stack := range stacks {
				stackIds = append(stackIds, stack.ID.String())
			}
			err = h.syncKeycloakWithClusterAdminPermission(r.Context(), organizationId, stackIds, []model.User{user})
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
//	@Security		JWT
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
//	@Security		JWT
func (h RoleHandler) Admin_GetTksRole(w http.ResponseWriter, r *http.Request) {
	// Same as GetTksRole

	vars := mux.Vars(r)
	var organizationId, roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		roleId = v
	}
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
	} else {
		organizationId = v
	}

	// get role
	role, err := h.roleUsecase.GetTksRole(r.Context(), organizationId, roleId)
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

// IsRoleNameExisted godoc
//
//	@Tags			Roles
//	@Summary		Check whether the role name exists
//	@Description	Check whether the role name exists
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			roleName		path		string	true	"Role Name"
//	@Success		200				{object}	domain.CheckRoleNameResponse
//	@Router			/organizations/{organizationId}/roles/{roleName}/existence [get]
//	@Security		JWT
func (h RoleHandler) IsRoleNameExisted(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var organizationId, roleName string
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}
	if v, ok := vars["roleName"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleName = v
	}

	// check role name exist
	isExist, err := h.roleUsecase.IsRoleNameExisted(r.Context(), organizationId, roleName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckRoleNameResponse
	out.IsExist = isExist

	// response
	ResponseJSON(w, r, http.StatusOK, out)
}

// AppendUsersToRole godoc
//
//	@Tags			Roles
//	@Summary		Append Users To Role
//	@Description	Append Users To Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string							true	"Organization ID"
//	@Param			roleId			path	string							true	"Role ID"
//	@Param			body			body	domain.AppendUsersToRoleRequest	true	"Append Users To Role Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId}/users [post]
//	@Security		JWT
func (h RoleHandler) AppendUsersToRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var organizationId, roleId string
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// request body
	input := domain.AppendUsersToRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	for _, userUuid := range input.Users {
		originUser, err := h.userUsecase.Get(r.Context(), userUuid)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		role, err := h.roleUsecase.GetTksRole(r.Context(), organizationId, roleId)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		originUser.Roles = append(originUser.Roles, *role)

		if _, err := h.userUsecase.UpdateByAccountIdByAdmin(r.Context(), originUser); err != nil {
			ErrorJSON(w, r, err)
			return
		}

		// Sync ClusterAdmin Permission to Keycloak
		stacks, err := h.stackUsecease.Fetch(r.Context(), organizationId, nil)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		stackIds := make([]string, 0)
		for _, stack := range stacks {
			stackIds = append(stackIds, stack.ID.String())
		}
		err = h.syncKeycloakWithClusterAdminPermission(r.Context(), organizationId, stackIds, []model.User{*originUser})
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// RemoveUsersFromRole godoc
//
//	@Tags			Roles
//	@Summary		Remove Users From Role
//	@Description	Remove Users From Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string								true	"Organization ID"
//	@Param			roleId			path	string								true	"Role ID"
//	@Param			body			body	domain.RemoveUsersFromRoleRequest	true	"Remove Users From Role Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId}/users [delete]
//	@Security		JWT
func (h RoleHandler) RemoveUsersFromRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var organizationId, roleId string
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// request body
	input := domain.RemoveUsersFromRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	for _, userUuid := range input.Users {
		originUser, err := h.userUsecase.Get(r.Context(), userUuid)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		role, err := h.roleUsecase.GetTksRole(r.Context(), organizationId, roleId)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		for i, r := range originUser.Roles {
			if r.ID == role.ID {
				originUser.Roles = append(originUser.Roles[:i], originUser.Roles[i+1:]...)
				break
			}
		}

		if _, err := h.userUsecase.UpdateByAccountIdByAdmin(r.Context(), originUser); err != nil {
			ErrorJSON(w, r, err)
			return
		}

		// Sync ClusterAdmin Permission to Keycloak
		stacks, err := h.stackUsecease.Fetch(r.Context(), organizationId, nil)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		stackIds := make([]string, 0)
		for _, stack := range stacks {
			stackIds = append(stackIds, stack.ID.String())
		}
		err = h.syncKeycloakWithClusterAdminPermission(r.Context(), organizationId, stackIds, []model.User{*originUser})
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetUsersInRoleId godoc
//
//	@Tags			Roles
//	@Summary		Get Users By Role ID
//	@Description	Get Users By Role ID
//	@Produce		json
//	@Param			organizationId	path		string		true	"Organization ID"
//	@Param			roleId			path		string		true	"Role ID"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.GetUsersInRoleIdResponse
//	@Router			/organizations/{organizationId}/roles/{roleId}/users [get]
//	@Security		JWT
func (h RoleHandler) GetUsersInRoleId(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var organizationId, roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	users, err := h.userUsecase.ListUsersByRole(r.Context(), organizationId, roleId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetUsersInRoleIdResponse
	out.Users = make([]domain.SimpleUserResponse, len(*users))
	for i, user := range *users {
		out.Users[i] = domain.SimpleUserResponse{
			ID:         user.ID.String(),
			AccountId:  user.AccountId,
			Name:       user.Name,
			Email:      user.Email,
			Department: user.Department,
		}
	}

	if err := serializer.Map(r.Context(), *pg, &out.Pagination); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// syncKeycloakWithClusterAdminPermission sync the permissions with Keycloak
// 1. Get all roles assigned to the user and merge the permissions
// 2. Then get the cluster admin permissions for the stack
// 3. Finally, sync the permissions with Keycloak
func (h RoleHandler) syncKeycloakWithClusterAdminPermission(ctx context.Context, organizationId string, clusterIds []string, users []model.User) error {
	for _, user := range users {
		// 1-step
		// Merge the permissions of the userUuid
		var permissionSets []*model.PermissionSet
		for _, role := range user.Roles {
			permissionSet, err := h.permissionUsecase.GetPermissionSetByRoleId(ctx, role.ID)
			if err != nil {
				return err
			}
			permissionSets = append(permissionSets, permissionSet)
		}
		mergedPermissionSet := h.permissionUsecase.MergePermissionWithOrOperator(ctx, permissionSets...)

		// 2-step
		// Then get the cluster admin permissions for the stack
		var targetEdgePermissions []*model.Permission
		// filter function
		f := func(permission model.Permission) bool {
			if permission.Parent != nil && permission.Parent.Key == model.MiddleClusterAccessControlKey {
				return true
			}
			return false
		}
		edgePermissions := model.GetEdgePermission(mergedPermissionSet.Stack, targetEdgePermissions, &f)

		// 3-step
		//  sync the permissions with Keycloak
		for _, clusterId := range clusterIds {
			if len(edgePermissions) > 0 {
				var err error
				for _, edgePermission := range edgePermissions {
					switch edgePermission.Key {
					case model.OperationCreate:
						err = h.permissionUsecase.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-create", *edgePermission.IsAllowed)
					case model.OperationRead:
						err = h.permissionUsecase.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-read", *edgePermission.IsAllowed)
					case model.OperationUpdate:
						err = h.permissionUsecase.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-update", *edgePermission.IsAllowed)
					case model.OperationDelete:
						err = h.permissionUsecase.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-delete", *edgePermission.IsAllowed)
					}
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
