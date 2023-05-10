package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
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
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	ctx := r.Context()
	var organization domain.Organization
	if err = domain.Map(input, &organization); err != nil {
		log.Error(err)
	}

	organizationId, err := h.usecase.Create(ctx, &organization)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}
	organization.ID = organizationId
	// Admin user 생성
	_, err = h.userUsecase.CreateAdmin(organizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateOrganizationResponse
	if err = domain.Map(organization, &out); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetOrganizations godoc
// @Tags Organizations
// @Summary Get organization list
// @Description Get organization list
// @Accept json
// @Produce json
// @Success 200 {object} []domain.ListOrganizationBody
// @Router /organizations [get]
// @Security     JWT
func (h *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.usecase.Fetch()
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}

	var out domain.ListOrganizationResponse
	out.Organizations = make([]domain.ListOrganizationBody, len(*organizations))

	for i, organization := range *organizations {
		if err = domain.Map(organization, &out.Organizations[i]); err != nil {
			log.Error(err)
		}

		log.Info(organization)
	}

	ResponseJSON(w, http.StatusOK, out)
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	organization, err := h.usecase.Get(organizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}

		ErrorJSON(w, err)
		return
	}

	log.Info("1")
	log.Info(organization)

	var out domain.GetOrganizationResponse
	if err = domain.Map(organization, &out.Organization); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token")))
		return
	}

	err := h.userUsecase.DeleteAll(r.Context(), organizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}

	// organization 삭제
	err = h.usecase.Delete(organizationId, token)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	input := domain.UpdateOrganizationRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	organization, err := h.usecase.Update(organizationId, input)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}
		ErrorJSON(w, err)
		return
	}

	var out domain.UpdateOrganizationResponse
	if err = domain.Map(organization, &out); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	input := domain.UpdatePrimaryClusterRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.UpdatePrimaryClusterId(organizationId, input.PrimaryClusterId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}
