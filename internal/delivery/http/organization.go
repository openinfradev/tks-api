package http

import (
	"encoding/json"
	"fmt"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type OrganizationHandler struct {
	organizationUsecase usecase.IOrganizationUsecase
	userUsecase         usecase.IUserUsecase
}

func NewOrganizationHandler(o usecase.IOrganizationUsecase, u usecase.IUserUsecase) *OrganizationHandler {
	return &OrganizationHandler{
		organizationUsecase: o,
		userUsecase:         u,
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
	userId, _ := GetSession(r)

	input := domain.CreateOrganizationRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		ErrorJSON(w, "token not found", http.StatusInternalServerError)
		return
	}

	organizationId, err := h.organizationUsecase.Create(domain.Organization{
		Name:        input.Name,
		Creator:     userId,
		Description: input.Description,
	}, token)
	if err != nil {
		log.Error("Failed to create organization err : ", err)
		//h.AddHistory(r, response.GetOrganizationId(), "organization", fmt.Sprintf("프로젝트 [%s]를 생성하는데 실패했습니다.", input.Name))
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Admin user 생성
	_, err = h.userUsecase.CreateAdmin(organizationId)
	if err != nil {
		log.Error("Failed to create user err : ", err)
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
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
	ResponseJSON(w, out, "", http.StatusOK)

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
	log.Info("GetOrganization")
	organizations, err := h.organizationUsecase.Fetch()
	if err != nil {
		log.Error("Failed to get organizations err : ", err)
		InternalServerError(w, err)
		return
	}

	var out struct {
		Organizations []domain.Organization `json:"organizations"`
	}

	out.Organizations = organizations

	ResponseJSON(w, out, "", http.StatusOK)
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
		ErrorJSON(w, fmt.Sprintf("Invalid input"), http.StatusBadRequest)
		return
	}

	organization, err := h.organizationUsecase.Get(organizationId)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		Organization domain.Organization `json:"organization"`
	}

	out.Organization = organization

	ResponseJSON(w, out, "", http.StatusOK)
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
		ErrorJSON(w, fmt.Sprintf("Invalid input %s", organizationId), http.StatusBadRequest)
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		http.Error(w, "token not found", http.StatusBadRequest)
		return
	}

	res, err := h.organizationUsecase.Delete(organizationId, token)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Failed to delete organization err : %s", err), http.StatusBadRequest)
		return
	}

	ResponseJSON(w, res, "", http.StatusOK)
}
