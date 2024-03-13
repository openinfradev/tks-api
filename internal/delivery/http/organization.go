package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type OrganizationHandler struct {
	usecase           usecase.IOrganizationUsecase
	userUsecase       usecase.IUserUsecase
	roleUsecase       usecase.IRoleUsecase
	permissionUsecase usecase.IPermissionUsecase
}

func NewOrganizationHandler(u usecase.Usecase) *OrganizationHandler {
	return &OrganizationHandler{
		usecase:           u.Organization,
		userUsecase:       u.User,
		roleUsecase:       u.Role,
		permissionUsecase: u.Permission,
	}
}

// CreateOrganization godoc
//
//	@Tags			Organizations
//	@Summary		Create organization
//	@Description	Create organization
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateOrganizationRequest	true	"create organization request"
//	@Success		200		{object}	object
//	@Router			/organizations [post]
//	@Security		JWT
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateOrganizationRequest{}

	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var organization model.Organization
	if err = serializer.Map(input, &organization); err != nil {
		log.Error(r.Context(), err)
	}

	organizationId, err := h.usecase.Create(ctx, &organization)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	organization.ID = organizationId

	// Role 생성
	adminRole := model.Role{
		OrganizationID: organizationId,
		Name:           "admin",
		Description:    "admin",
		Type:           string(domain.RoleTypeTks),
	}
	adminRoleId, err := h.roleUsecase.CreateTksRole(r.Context(), &adminRole)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	userRole := model.Role{
		OrganizationID: organizationId,
		Name:           "user",
		Description:    "user",
		Type:           string(domain.RoleTypeTks),
	}
	userRoleId, err := h.roleUsecase.CreateTksRole(r.Context(), &userRole)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	// Permission 생성
	adminPermissionSet := h.permissionUsecase.GetAllowedPermissionSet(r.Context())
	h.permissionUsecase.SetRoleIdToPermissionSet(r.Context(), adminRoleId, adminPermissionSet)
	err = h.permissionUsecase.CreatePermissionSet(r.Context(), adminPermissionSet)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	userPermissionSet := h.permissionUsecase.GetUserPermissionSet(r.Context())
	h.permissionUsecase.SetRoleIdToPermissionSet(r.Context(), userRoleId, userPermissionSet)
	err = h.permissionUsecase.CreatePermissionSet(r.Context(), userPermissionSet)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	// Admin user 생성
	admin, err := h.userUsecase.CreateAdmin(r.Context(), organizationId, input.AdminAccountId, input.AdminName, input.AdminEmail)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.ChangeAdminId(r.Context(), organizationId, admin.ID)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	organization.AdminId = &admin.ID

	var out domain.CreateOrganizationResponse
	if err = serializer.Map(organization, &out); err != nil {
		log.Error(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetOrganizations godoc
//
//	@Tags			Organizations
//	@Summary		Get organization list
//	@Description	Get organization list
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	[]domain.ListOrganizationResponse
//	@Router			/organizations [get]
//	@Security		JWT
func (h *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	organizations, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListOrganizationResponse
	out.Organizations = make([]domain.OrganizationResponse, len(*organizations))

	for i, organization := range *organizations {
		if err = serializer.Map(organization, &out.Organizations[i]); err != nil {
			log.Error(r.Context(), err)
		}

		log.Info(r.Context(), organization)
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetOrganization godoc
//
//	@Tags			Organizations
//	@Summary		Get organization detail
//	@Description	Get organization detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetOrganizationResponse
//	@Router			/organizations/{organizationId} [get]
//	@Security		JWT
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	organization, err := h.usecase.Get(r.Context(), organizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}
	var out domain.GetOrganizationResponse
	if err = serializer.Map(organization, &out.Organization); err != nil {
		log.Error(r.Context(), err)
	}

	out.Organization.StackTemplates = make([]domain.SimpleStackTemplateResponse, len(organization.StackTemplates))
	for i, stackTemplate := range organization.StackTemplates {
		if err = serializer.Map(stackTemplate, &out.Organization.StackTemplates[i]); err != nil {
			log.ErrorWithContext(r.Context(), err)
		}
		err := json.Unmarshal(stackTemplate.Services, &out.Organization.StackTemplates[i].Services)
		if err != nil {
			log.ErrorWithContext(r.Context(), err)
		}
	}
	out.Organization.PolicyTemplates = make([]domain.SimplePolicyTemplateResponse, len(organization.PolicyTemplates))
	for i, policyTemplate := range organization.PolicyTemplates {
		if err = serializer.Map(policyTemplate, &out.Organization.PolicyTemplates[i]); err != nil {
			log.ErrorWithContext(r.Context(), err)
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// DeleteOrganization godoc
//
//	@Tags			Organizations
//	@Summary		Delete organization
//	@Description	Delete organization
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.DeleteOrganizationResponse
//	@Router			/organizations/{organizationId} [delete]
//	@Security		JWT
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", ""))
		return
	}

	err := h.userUsecase.DeleteAll(r.Context(), organizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	// organization 삭제
	err = h.usecase.Delete(r.Context(), organizationId, token)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	out := domain.DeleteOrganizationResponse{
		ID: organizationId,
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateOrganization godoc
//
//	@Tags			Organizations
//	@Summary		Update organization detail
//	@Description	Update organization detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"organizationId"
//	@Param			body			body		domain.UpdateOrganizationRequest	true	"update organization request"
//	@Success		200				{object}	domain.UpdateOrganizationResponse
//	@Router			/organizations/{organizationId} [put]
//	@Security		JWT
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.UpdateOrganizationRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	organization, err := h.usecase.Update(r.Context(), organizationId, input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateOrganizationResponse
	if err = serializer.Map(organization, &out); err != nil {
		log.Error(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePrimaryCluster godoc
//
//	@Tags			Organizations
//	@Summary		Update primary cluster
//	@Description	Update primary cluster
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"organizationId"
//	@Param			body			body		domain.UpdatePrimaryClusterRequest	true	"update primary cluster request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/primary-cluster [patch]
//	@Security		JWT
func (h *OrganizationHandler) UpdatePrimaryCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.UpdatePrimaryClusterRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.UpdatePrimaryClusterId(r.Context(), organizationId, input.PrimaryClusterId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
