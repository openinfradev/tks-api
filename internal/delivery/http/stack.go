package http

import (
	"encoding/json"
	"fmt"
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

type StackHandler struct {
	usecase usecase.IStackUsecase
}

func NewStackHandler(h usecase.Usecase) *StackHandler {
	return &StackHandler{
		usecase: h.Stack,
	}
}

// CreateStack godoc
//
//	@Tags			Stacks
//	@Summary		Create Stack
//	@Description	Create Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			body			body		domain.CreateStackRequest	true	"create cloud setting request"
//	@Success		200				{object}	domain.CreateStackResponse
//	@Router			/organizations/{organizationId}/stacks [post]
//	@Security		JWT
func (h *StackHandler) CreateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.CreateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.Stack
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.OrganizationId = organizationId
	stackId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.CreateStackResponse{
		ID: stackId.String(),
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

func (h *StackHandler) InstallStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "S_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.Install(r.Context(), domain.StackId(stackId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetStack godoc
//
//	@Tags			Stacks
//	@Summary		Get Stacks
//	@Description	Get Stacks
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			limit			query		string	false	"pageSize"
//	@Param			page			query		string	false	"pageNumber"
//	@Param			soertColumn		query		string	false	"sortColumn"
//	@Param			sortOrder		query		string	false	"sortOrder"
//	@Success		200				{object}	domain.GetStacksResponse
//	@Router			/organizations/{organizationId}/stacks [get]
//	@Security		JWT
func (h *StackHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	stacks, err := h.usecase.Fetch(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStacksResponse
	out.Stacks = make([]domain.StackResponse, len(stacks))
	for i, stack := range stacks {
		if err := serializer.Map(r.Context(), stack, &out.Stacks[i]); err != nil {
			log.Info(r.Context(), err)
		}

		if err := serializer.Map(r.Context(), stack.CreatedAt, &out.Stacks[i].CreatedAt); err != nil {
			log.Info(r.Context(), err)
		}

		err = json.Unmarshal(stack.StackTemplate.Services, &out.Stacks[i].StackTemplate.Services)
		if err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStack godoc
//
//	@Tags			Stacks
//	@Summary		Get Stack
//	@Description	Get Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	domain.GetStackResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId} [get]
//	@Security		JWT
func (h *StackHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	stack, err := h.usecase.Get(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackResponse
	if err := serializer.Map(r.Context(), stack, &out.Stack); err != nil {
		log.Info(r.Context(), err)
	}

	err = json.Unmarshal(stack.StackTemplate.Services, &out.Stack.StackTemplate.Services)
	if err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackStatus godoc
//
//	@Tags			Stacks
//	@Summary		Get Stack Status
//	@Description	Get Stack Status
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	domain.GetStackStatusResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/status [get]
//	@Security		JWT
func (h *StackHandler) GetStackStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	steps, status, err := h.usecase.GetStepStatus(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackStatusResponse
	out.StepStatus = make([]domain.StackStepStatus, len(steps))
	for i, step := range steps {
		if err := serializer.Map(r.Context(), step, &out.StepStatus[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}
	out.StackStatus = status

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateStack godoc
//
//	@Tags			Stacks
//	@Summary		Update Stack
//	@Description	Update Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			stackId			path		string						true	"stackId"
//	@Param			body			body		domain.UpdateStackRequest	true	"Update cloud setting request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId} [put]
//	@Security		JWT
func (h *StackHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.UpdateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.Stack
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.ID = stackId
	dto.OrganizationId = organizationId

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)

}

// DeleteStack godoc
//
//	@Tags			Stacks
//	@Summary		Delete Stack
//	@Description	Delete Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId} [delete]
//	@Security		JWT
func (h *StackHandler) DeleteStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	var dto model.Stack
	dto.ID = domain.StackId(strId)
	dto.OrganizationId = organizationId

	err := h.usecase.Delete(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// CheckStackName godoc
//
//	@Tags			Stacks
//	@Summary		Check name for stack
//	@Description	Check name for stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Param			name			path		string	true	"name"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/name/{name}/existence [GET]
//	@Security		JWT
func (h *StackHandler) CheckStackName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name"), "S_INVALID_STACK_NAME", ""))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(r.Context(), organizationId, name)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckStackNameResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackKubeConfig godoc
//
//	@Tags			Stacks
//	@Summary		Get KubeConfig by stack
//	@Description	Get KubeConfig by stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetStackKubeConfigResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/kube-config [get]
//	@Security		JWT
func (h *StackHandler) GetStackKubeConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	kubeConfig, err := h.usecase.GetKubeConfig(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out = domain.GetStackKubeConfigResponse{
		KubeConfig: kubeConfig,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// SetFavorite godoc
//
//	@Tags			Stacks
//	@Summary		Set favorite stack
//	@Description	Set favorite stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/favorite [post]
//	@Security		JWT
func (h *StackHandler) SetFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.SetFavorite(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteFavorite godoc
//
//	@Tags			Stacks
//	@Summary		Delete favorite stack
//	@Description	Delete favorite stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/favorite [delete]
//	@Security		JWT
func (h *StackHandler) DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.DeleteFavorite(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}
