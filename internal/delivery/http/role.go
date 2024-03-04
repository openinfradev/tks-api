package http

import (
	"github.com/gorilla/mux"
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
// @Tags Role
// @Summary Create Tks Role
// @Description Create Tks Role
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param body body domain.CreateTksRoleRequest true "Create Tks Role Request"
// @Success 200 {object} domain.CreateTksRoleResponse
// @Router /organizations/{organizationId}/roles [post]

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
	dto := domain.Role{
		OrganizationID: organizationId,
		Name:           input.Name,
		Description:    input.Description,
		Type:           string(domain.RoleTypeTks),
	}

	// create role
	var roleId string
	if roleId, err = h.roleUsecase.CreateTksRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// create permission
	defaultPermissionSet := domain.NewDefaultPermissionSet()
	h.permissionUsecase.SetRoleIdToPermissionSet(roleId, defaultPermissionSet)
	err = h.permissionUsecase.CreatePermissionSet(defaultPermissionSet)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, domain.CreateTksRoleResponse{ID: roleId})
}

// ListTksRoles godoc
// @Tags Role
// @Summary List Tks Roles
// @Description List Tks Roles
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Success 200 {object} domain.ListTksRoleResponse
// @Router /organizations/{organizationId}/roles [get]
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
	roles, err := h.roleUsecase.ListTksRoles(organizationId, pg)
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

	if err := serializer.Map(*pg, &out.Pagination); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	// response
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetTksRole godoc
// @Tags Role
// @Summary Get Tks Role
// @Description Get Tks Role
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} domain.GetTksRoleResponse
// @Router /organizations/{organizationId}/roles/{roleId} [get]

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
	role, err := h.roleUsecase.GetTksRole(roleId)
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
// @Tags Role
// @Summary Delete Tks Role
// @Description Delete Tks Role
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param roleId path string true "Role ID"
// @Success 200
// @Router /organizations/{organizationId}/roles/{roleId} [delete]

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
	if err := h.roleUsecase.DeleteTksRole(roleId); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// UpdateTksRole godoc
// @Tags Role
// @Summary Update Tks Role
// @Description Update Tks Role
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param roleId path string true "Role ID"
// @Param body body domain.UpdateTksRoleRequest true "Update Tks Role Request"
// @Success 200
// @Router /organizations/{organizationId}/roles/{roleId} [put]

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
	dto := domain.Role{
		ID:          roleId,
		Description: input.Description,
	}

	// update role
	if err := h.roleUsecase.UpdateTksRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}
