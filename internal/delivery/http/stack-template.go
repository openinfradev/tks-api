package http

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type StackTemplateHandler struct {
	usecase usecase.IStackTemplateUsecase
}

func NewStackTemplateHandler(h usecase.IStackTemplateUsecase) *StackTemplateHandler {
	return &StackTemplateHandler{
		usecase: h,
	}
}

// CreateStackTemplate godoc
// @Tags StackTemplates
// @Summary Create StackTemplate 'NOT IMPLEMENTED'
// @Description Create StackTemplate
// @Accept json
// @Produce json
// @Param body body domain.CreateStackTemplateRequest true "create stack template request"
// @Success 200 {object} domain.CreateStackTemplateResponse
// @Router /stack-templates [post]
// @Security     JWT
func (h *StackTemplateHandler) CreateStackTemplate(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, fmt.Errorf("Need implentation"))
}

// GetStackTemplate godoc
// @Tags StackTemplates
// @Summary Get StackTemplates
// @Description Get StackTemplates
// @Accept json
// @Produce json
// @Success 200 {object} domain.GetStackTemplatesResponse
// @Router /stack-templates [get]
// @Security     JWT
func (h *StackTemplateHandler) GetStackTemplates(w http.ResponseWriter, r *http.Request) {
	stackTemplates, err := h.usecase.Fetch()
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetStackTemplatesResponse
	out.StackTemplates = make([]domain.StackTemplateResponse, len(stackTemplates))
	for i, stackTemplate := range stackTemplates {
		if err := domain.Map(stackTemplate, &out.StackTemplates[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetStackTemplate godoc
// @Tags StackTemplates
// @Summary Get StackTemplate
// @Description Get StackTemplate
// @Accept json
// @Produce json
// @Param stackTemplateId path string true "stackTemplateId"
// @Success 200 {object} domain.GetStackTemplateResponse
// @Router /stack-templates/{stackTemplateId} [get]
// @Security     JWT
func (h *StackTemplateHandler) GetStackTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackTemplateId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackTemplateId")))
		return
	}

	stackTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	stackTemplate, err := h.usecase.Get(stackTemplateId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetStackTemplateResponse
	if err := domain.Map(stackTemplate, &out.StackTemplate); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)

}

// UpdateStackTemplate godoc
// @Tags StackTemplates
// @Summary Update StackTemplate 'NOT IMPLEMENTED'
// @Description Update StackTemplate
// @Accept json
// @Produce json
// @Param body body domain.UpdateStackTemplateRequest true "Update stack template request"
// @Success 200 {object} nil
// @Router /stack-templates/{stackTemplateId} [put]
// @Security     JWT
func (h *StackTemplateHandler) UpdateStackTemplate(w http.ResponseWriter, r *http.Request) {
	/*
		vars := mux.Vars(r)
		strId, ok := vars["stackTemplateId"]
		if !ok {
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackTemplateId")))
			return
		}

		stackTemplateId, err := uuid.Parse(strId)
		if err != nil {
			ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
			return
		}

		var dto domain.StackTemplate
		if err := domain.Map(r, &dto); err != nil {
			log.Info(err)
		}
		dto.ID = stackTemplateId

		err = h.usecase.Update(r.Context(), dto)
		if err != nil {
			ErrorJSON(w, err)
			return
		}
	*/

	ErrorJSON(w, fmt.Errorf("Need implentation"))
}

// DeleteStackTemplate godoc
// @Tags StackTemplates
// @Summary Delete StackTemplate 'NOT IMPLEMENTED'
// @Description Delete StackTemplate
// @Accept json
// @Produce json
// @Param stackTemplateId path string true "stackTemplateId"
// @Success 200 {object} nil
// @Router /stack-templates/{stackTemplateId} [delete]
// @Security     JWT
func (h *StackTemplateHandler) DeleteStackTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["stackTemplateId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackTemplateId")))
		return
	}

	ErrorJSON(w, fmt.Errorf("Need implentation"))
}
