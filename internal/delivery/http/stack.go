package http

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
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
// @Param body body domain.CreateStackRequest true "create cloud setting request"
// @Success 200 {object} domain.CreateStackResponse
// @Router /stacks [post]
// @Security     JWT
func (h *StackHandler) CreateStack(w http.ResponseWriter, r *http.Request) {
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

	stackId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateStackResponse
	out.ID = stackId.String()

	ResponseJSON(w, http.StatusOK, out)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get Stacks
// @Description Get Stacks
// @Accept json
// @Produce json
// @Param all query string false "show all organizations"
// @Success 200 {object} domain.GetStacksResponse
// @Router /stacks [get]
// @Security     JWT
func (h *StackHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	user, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token")))
		return
	}

	urlParams := r.URL.Query()
	showAll := urlParams.Get("all")

	// [TODO REFACTORING] Privileges and Filtering
	if showAll == "true" {
		ErrorJSON(w, httpErrors.NewUnauthorizedError(fmt.Errorf("Your token does not have permission to see all organizations.")))
		return
	}

	stacks, err := h.usecase.Fetch(user.GetOrganizationId())
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
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get Stack
// @Description Get Stack
// @Accept json
// @Produce json
// @Param stackId path string true "stackId"
// @Success 200 {object} domain.GetStackResponse
// @Router /stacks/{stackId} [get]
// @Security     JWT
func (h *StackHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	stackId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	stack, err := h.usecase.Get(stackId)
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
// @Param body body domain.UpdateStackRequest true "Update cloud setting request"
// @Success 200 {object} nil
// @Router /stacks/{stackId} [put]
// @Security     JWT
func (h *StackHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	cloudSeetingId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	input := domain.UpdateStackRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.Stack
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}
	dto.ID = cloudSeetingId

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
// @Param body body domain.DeleteStackRequest true "Delete cloud setting request"
// @Param stackId path string true "stackId"
// @Success 200 {object} nil
// @Router /stacks/{stackId} [delete]
// @Security     JWT
func (h *StackHandler) DeleteStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId")))
		return
	}

	parsedId, err := uuid.Parse(stackId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid")))
		return
	}

	input := domain.DeleteStackRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.Stack
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}
	dto.ID = parsedId

	err = h.usecase.Delete(r.Context(), dto)
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
// @Param name path string true "name"
// @Success 200 {object} nil
// @Router /stacks/name/{name}/existence [GET]
// @Security     JWT
func (h *StackHandler) CheckStackName(w http.ResponseWriter, r *http.Request) {
	user, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token")))
		return
	}

	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name")))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(user.GetOrganizationId(), name)
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
