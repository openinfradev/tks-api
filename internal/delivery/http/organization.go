package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type OrganizationHandler struct {
	usecase     usecase.IOrganizationUsecase
	userUsecase usecase.IUserUsecase
}

func NewOrganizationHandler(o usecase.IOrganizationUsecase, u usecase.IUserUsecase) *OrganizationHandler {
	return &OrganizationHandler{
		usecase:     o,
		userUsecase: u,
	}
}

// CreateOrganization godoc
// @Tags Organizations
// @Summary Create organization
// @Description Create organization
// @Accept json
// @Produce json
// @Param body body domain.CreateOrganizationRequest true "create organization request"
// @Success 200 {object} object
// @Router /organizations [post]
// @Security     JWT
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateOrganizationRequest{}

	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var organization domain.Organization
	if err = domain.Map(input, &organization); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	organizationId, err := h.usecase.Create(ctx, &organization)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	organization.ID = organizationId
	// Admin user 생성
	_, err = h.userUsecase.CreateAdmin(organizationId, input.Email)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateOrganizationResponse
	if err = domain.Map(organization, &out); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetOrganizations godoc
// @Tags Organizations
// @Summary Get organization list
// @Description Get organization list
// @Accept json
// @Produce json
// @Param limit query string false "pageSize"
// @Param page query string false "pageNumber"
// @Param soertColumn query string false "sortColumn"
// @Param sortOrder query string false "sortOrder"
// @Param filters query []string false "filters"
// @Success 200 {object} []domain.ListOrganizationBody
// @Router /organizations [get]
// @Security     JWT
func (h *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	organizations, err := h.usecase.Fetch(pg)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListOrganizationResponse
	out.Organizations = make([]domain.ListOrganizationBody, len(*organizations))

	for i, organization := range *organizations {
		if err = domain.Map(organization, &out.Organizations[i]); err != nil {
			log.ErrorWithContext(r.Context(), err)
		}

		log.InfoWithContext(r.Context(), organization)
	}

	if err := domain.Map(*pg, &out.Pagination); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetOrganization godoc
// @Tags Organizations
// @Summary Get organization detail
// @Description Get organization detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.GetOrganizationResponse
// @Router /organizations/{organizationId} [get]
// @Security     JWT
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	organization, err := h.usecase.Get(organizationId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}
	var out domain.GetOrganizationResponse
	if err = domain.Map(organization, &out.Organization); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// DeleteOrganization godoc
// @Tags Organizations
// @Summary Delete organization
// @Description Delete organization
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.Organization
// @Router /organizations/{organizationId} [delete]
// @Security     JWT
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	// organization 삭제
	err = h.usecase.Delete(organizationId, token)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// UpdateOrganization godoc
// @Tags Organizations
// @Summary Update organization detail
// @Description Update organization detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param body body domain.UpdateOrganizationRequest true "update organization request"
// @Success 200 {object} domain.UpdateOrganizationResponse
// @Router /organizations/{organizationId} [put]
// @Security     JWT
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

	organization, err := h.usecase.Update(organizationId, input)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateOrganizationResponse
	if err = domain.Map(organization, &out); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePrimaryCluster godoc
// @Tags Organizations
// @Summary Update primary cluster
// @Description Update primary cluster
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param body body domain.UpdatePrimaryClusterRequest true "update primary cluster request"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/primary-cluster [patch]
// @Security     JWT
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

	err = h.usecase.UpdatePrimaryClusterId(organizationId, input.PrimaryClusterId)
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
