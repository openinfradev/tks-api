package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type StackHandler struct {
	usecase usecase.IStackUsecase
}

func NewStackHandler(h usecase.IStackUsecase) *StackHandler {
	return &StackHandler{
		usecase: h,
	}
}

// CreateStack godoc
// @Tags Stacks
// @Summary Create Stack
// @Description Create Stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param body body domain.CreateStackRequest true "create cloud setting request"
// @Success 200 {object} domain.CreateStackResponse
// @Router /organizations/{organizationId}/stacks [post]
// @Security     JWT
func (h *StackHandler) CreateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	input := domain.CreateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.Stack
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}
	if err = domain.Map(input, &dto.Conf); err != nil {
		log.Info(err)
	}
	dto.OrganizationId = organizationId

	stackId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	out := domain.CreateStackResponse{
		ID: stackId.String(),
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get Stacks
// @Description Get Stacks
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.GetStacksResponse
// @Router /organizations/{organizationId}/stacks [get]
// @Security     JWT
func (h *StackHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	log.Debug("[TODO] organization check", organizationId)

	stacks, err := h.usecase.Fetch(organizationId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetStacksResponse
	out.Stacks = make([]domain.StackResponse, len(stacks))
	for i, stack := range stacks {
		if err := domain.Map(stack, &out.Stacks[i]); err != nil {
			log.Info(err)
			continue
		}
		log.Info(out.Stacks[i])
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get Stack
// @Description Get Stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param stackId path string true "stackId"
// @Success 200 {object} domain.GetStackResponse
// @Router /organizations/{organizationId}/stacks/{stackId} [get]
// @Security     JWT
func (h *StackHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	log.Debug("[TODO] organization check", organizationId)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	stack, err := h.usecase.Get(domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetStackResponse
	if err := domain.Map(stack, &out.Stack); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateStack godoc
// @Tags Stacks
// @Summary Update Stack
// @Description Update Stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param stackId path string true "stackId"
// @Param body body domain.UpdateStackRequest true "Update cloud setting request"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/stacks/{stackId} [put]
// @Security     JWT
func (h *StackHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	log.Debug("[TODO] organization check", organizationId)

	input := domain.UpdateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.Stack
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}
	dto.ID = stackId
	dto.OrganizationId = organizationId

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)

}

// DeleteStack godoc
// @Tags Stacks
// @Summary Delete Stack
// @Description Delete Stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param stackId path string true "stackId"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/stacks/{stackId} [delete]
// @Security     JWT
func (h *StackHandler) DeleteStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	log.Debug("[TODO] organization check", organizationId)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	_, err := h.usecase.Get(domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// CheckStackName godoc
// @Tags Stacks
// @Summary Check name for stack
// @Description Check name for stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param stackId path string true "stackId"
// @Param name path string true "name"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/stacks/name/{name}/existence [GET]
// @Security     JWT
func (h *StackHandler) CheckStackName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name")))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(organizationId, name)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, err)
			return
		}
	}

	var out domain.CheckStackNameResponse
	out.Existed = exist

	ResponseJSON(w, http.StatusOK, out)
}

// GetStackKubeConfig godoc
// @Tags Stacks
// @Summary Get KubeConfig by stack
// @Description Get KubeConfig by stack
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param stackId path string true "organizationId"
// @Success 200 {object} domain.GetStackKubeConfigResponse
// @Router /organizations/{organizationId}/stacks/{stackId}/kube-config [get]
// @Security     JWT
func (h *StackHandler) GetStackKubeConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	kubeConfig, err := h.usecase.GetKubeConfig(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out = domain.GetStackKubeConfigResponse{
		KubeConfig: kubeConfig,
	}

	ResponseJSON(w, http.StatusOK, out)
}
