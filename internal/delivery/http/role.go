package http

import (
	"github.com/google/uuid"
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
	CreateProjectRole(w http.ResponseWriter, r *http.Request)

	ListTksRoles(w http.ResponseWriter, r *http.Request)
	ListProjectRoles(w http.ResponseWriter, r *http.Request)

	GetTksRole(w http.ResponseWriter, r *http.Request)
	GetProjectRole(w http.ResponseWriter, r *http.Request)

	DeleteTksRole(w http.ResponseWriter, r *http.Request)
	DeleteProjectRole(w http.ResponseWriter, r *http.Request)

	UpdateTksRole(w http.ResponseWriter, r *http.Request)
	UpdateProjectRole(w http.ResponseWriter, r *http.Request)
}

type RoleHandler struct {
	roleUsecase usecase.IRoleUsecase
}

func NewRoleHandler(roleUsecase usecase.IRoleUsecase) *RoleHandler {
	return &RoleHandler{
		roleUsecase: roleUsecase,
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
	dto := domain.TksRole{
		Role: domain.Role{
			OrganizationID: organizationId,
			Name:           input.Name,
			Description:    input.Description,
			Type:           string(domain.RoleTypeTks),
		},
	}

	if err := h.roleUsecase.CreateTksRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

}

// CreateTksRole godoc
// @Tags Role
// @Summary Create Project Role
// @Description Create Project Role
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param body body domain.CreateProjectRoleRequest true "Create Project Role Request"
// @Success 200 {object} domain.CreateProjectRoleResponse
// @Router /organizations/{organizationId}/projects/{projectId}/roles [post]

func (h RoleHandler) CreateProjectRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var organizationId, projectId string

	vars := mux.Vars(r)
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}
	if v, ok := vars["projectId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		projectId = v
	}

	// request body
	input := domain.CreateProjectRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// input to dto
	projectIdUuid, err := uuid.Parse(projectId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}
	dto := domain.ProjectRole{
		Role: domain.Role{
			OrganizationID: organizationId,
			Name:           input.Name,
			Description:    input.Description,
			Type:           string(domain.RoleTypeProject),
		},
		ProjectID: projectIdUuid,
	}

	if err := h.roleUsecase.CreateProjectRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}
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
	pg, err := pagination.NewPagination(&urlParams)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

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
			ID:             role.ID.String(),
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

// ListProjectRoles godoc
// @Tags Role
// @Summary List Project Roles
// @Description List Project Roles
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Success 200 {object} domain.ListProjectRoleResponse
// @Router /organizations/{organizationId}/projects/{projectId}/roles [get]

func (h RoleHandler) ListProjectRoles(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var projectId string

	vars := mux.Vars(r)
	if v, ok := vars["projectId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		projectId = v
	}

	// query parameter
	urlParams := r.URL.Query()
	pg, err := pagination.NewPagination(&urlParams)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	// list roles
	roles, err := h.roleUsecase.ListProjectRoles(projectId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListProjectRoleResponse
	out.Roles = make([]domain.GetProjectRoleResponse, len(roles))
	for i, role := range roles {
		out.Roles[i] = domain.GetProjectRoleResponse{
			ID:             role.RoleID.String(),
			Name:           role.Role.Name,
			OrganizationID: role.Role.OrganizationID,
			ProjectID:      role.ProjectID.String(),
			Description:    role.Role.Description,
			Creator:        role.Role.Creator.String(),
			CreatedAt:      role.Role.CreatedAt,
			UpdatedAt:      role.Role.UpdatedAt,
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
		ID:             role.ID.String(),
		Name:           role.Name,
		OrganizationID: role.OrganizationID,
		Description:    role.Description,
		Creator:        role.Creator.String(),
		CreatedAt:      role.CreatedAt,
		UpdatedAt:      role.UpdatedAt,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectRole godoc
// @Tags Role
// @Summary Get Project Role
// @Description Get Project Role
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} domain.GetProjectRoleResponse
// @Router /organizations/{organizationId}/projects/{projectId}/roles/{roleId} [get]

func (h RoleHandler) GetProjectRole(w http.ResponseWriter, r *http.Request) {
	// path parameter
	vars := mux.Vars(r)
	var roleId string
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// get role
	role, err := h.roleUsecase.GetProjectRole(roleId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	out := domain.GetProjectRoleResponse{
		ID:             role.RoleID.String(),
		Name:           role.Role.Name,
		OrganizationID: role.Role.OrganizationID,
		ProjectID:      role.ProjectID.String(),
		Description:    role.Role.Description,
		Creator:        role.Role.Creator.String(),
		CreatedAt:      role.Role.CreatedAt,
		UpdatedAt:      role.Role.UpdatedAt,
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

// DeleteProjectRole godoc
// @Tags Role
// @Summary Delete Project Role
// @Description Delete Project Role
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param roleId path string true "Role ID"
// @Success 200
// @Router /organizations/{organizationId}/projects/{projectId}/roles/{roleId} [delete]

func (h RoleHandler) DeleteProjectRole(w http.ResponseWriter, r *http.Request) {
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
	if err := h.roleUsecase.DeleteProjectRole(roleId); err != nil {
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
	roleIdUuid, err := uuid.Parse(roleId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}
	dto := domain.TksRole{
		Role: domain.Role{
			ID:          roleIdUuid,
			Name:        input.Name,
			Description: input.Description,
		},
	}

	// update role
	if err := h.roleUsecase.UpdateTksRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}

// UpdateProjectRole godoc
// @Tags Role
// @Summary Update Project Role
// @Description Update Project Role
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param roleId path string true "Role ID"
// @Param body body domain.UpdateProjectRoleRequest true "Update Project Role Request"
// @Success 200
// @Router /organizations/{organizationId}/projects/{projectId}/roles/{roleId} [put]

func (h RoleHandler) UpdateProjectRole(w http.ResponseWriter, r *http.Request) {
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
	input := domain.UpdateProjectRoleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// input to dto
	roleIdUuid, err := uuid.Parse(roleId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}
	dto := domain.ProjectRole{
		Role: domain.Role{
			ID:          roleIdUuid,
			Name:        input.Name,
			Description: input.Description,
		},
	}

	// update role
	if err := h.roleUsecase.UpdateProjectRole(&dto); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// response
	ResponseJSON(w, r, http.StatusOK, nil)
}
