package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/auth/request"
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

	err = h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get Stacks
// @Description Get Stacks
// @Accept json
// @Produce json
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
// @Param body body domain.UpdateStackRequest true "Update cloud setting request"
// @Success 200 {object} nil
// @Router /stacks/{stackId} [put]
// @Security     JWT
func (h *StackHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("Need implementaion")))
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
	ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("Need implementaion")))
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
	ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("Need implementaion")))

}
