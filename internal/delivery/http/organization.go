package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type OrganizationHandler struct {
	usecase usecase.IOrganizationUsecase
}

func NewOrganizationHandler(h usecase.IOrganizationUsecase) *OrganizationHandler {
	return &OrganizationHandler{
		usecase: h,
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
	_, userId, _ := GetSession(r)

	input := domain.CreateOrganizationRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	validate := validator.New()
	err = validate.Struct(input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewUnauthorizedError(fmt.Errorf("token not found")))
		return
	}

	organizationId, err := h.usecase.Create(domain.Organization{
		Name:        input.Name,
		Creator:     userId.String(),
		Description: input.Description,
	}, token)
	if err != nil {
		log.Error("Failed to create organization err : ", err)
		//h.AddHistory(r, response.GetOrganizationId(), "organization", fmt.Sprintf("프로젝트 [%s]를 생성하는데 실패했습니다.", input.Name))
		ErrorJSON(w, err)
		return
	}

	// add user to organizationUser
	/*
		err = h.Repository.AddUserInOrganization(userId, response.GetOrganizationId())
		if err != nil {
			h.AddHistory(r, response.GetOrganizationId(), "organization", fmt.Sprintf("프로젝트 [%s]를 생성하는데 실패했습니다.", input.Name))
			ErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/

	var out struct {
		OrganizationId string `json:"id"`
	}

	out.OrganizationId = organizationId

	//h.AddHistory(r, response.GetOrganizationId(), "organization", fmt.Sprintf("프로젝트 [%s]를 생성하였습니다.", out.OrganizationId))

	time.Sleep(time.Second * 5) // for test
	ResponseJSON(w, http.StatusOK, out)

}

// GetOrganizations godoc
// @Tags Organizations
// @Summary Get organization list
// @Description Get organization list
// @Accept json
// @Produce json
// @Success 200 {object} []domain.Organization
// @Router /organizations [get]
// @Security     JWT
func (h *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.usecase.Fetch()
	if err != nil {
		log.Error("Failed to get organizations err : ", err)
		ErrorJSON(w, err)
		return
	}

	var out struct {
		Organizations []domain.Organization `json:"organizations"`
	}

	out.Organizations = organizations

	ResponseJSON(w, http.StatusOK, out)
}

// GetOrganization godoc
// @Tags Organizations
// @Summary Get organization detail
// @Description Get organization detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.Organization
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
		ErrorJSON(w, err)
		return
	}

	var out struct {
		Organization domain.Organization `json:"organization"`
	}

	out.Organization = organization

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

	err := h.usecase.Delete(organizationId, token)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// GetOrganization godoc
// @Tags Organizations
// @Summary Update organization detail
// @Description Update organization detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.Organization
// @Router /organizations/{organizationId} [put]
// @Security     JWT
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	input := domain.UpdateOrganizationRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.Update(organizationId, input)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}
