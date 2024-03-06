package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type StackTemplateHandler struct {
	usecase usecase.IStackTemplateUsecase
}

func NewStackTemplateHandler(h usecase.Usecase) *StackTemplateHandler {
	return &StackTemplateHandler{
		usecase: h.StackTemplate,
	}
}

// CreateStackTemplate godoc
//
//	@Tags			StackTemplates
//	@Summary		Create StackTemplate 'NOT IMPLEMENTED'
//	@Description	Create StackTemplate
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateStackTemplateRequest	true	"create stack template request"
//	@Success		200		{object}	domain.CreateStackTemplateResponse
//	@Router			/admin/1.0/stack-templates [post]
//	@Security		JWT
func (h *StackTemplateHandler) CreateStackTemplate(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateStackTemplateRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.StackTemplate
	if err = serializer.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	id, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.CreateStackTemplateResponse{
		ID: id.String(),
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackTemplate godoc
//
//	@Tags			StackTemplates
//	@Summary		Get StackTemplates
//	@Description	Get StackTemplates
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetStackTemplatesResponse
//	@Router			/api/1.0/stack-templates [get]
//	@Security		JWT
func (h *StackTemplateHandler) GetStackTemplates(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	stackTemplates, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackTemplatesResponse
	out.StackTemplates = make([]domain.StackTemplateResponse, len(stackTemplates))
	for i, stackTemplate := range stackTemplates {
		if err := serializer.Map(stackTemplate, &out.StackTemplates[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
		}

		err := json.Unmarshal(stackTemplate.Services, &out.StackTemplates[i].Services)
		if err != nil {
			log.ErrorWithContext(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackTemplate godoc
//
//	@Tags			StackTemplates
//	@Summary		Get StackTemplate
//	@Description	Get StackTemplate
//	@Accept			json
//	@Produce		json
//	@Param			stackTemplateId	path		string	true	"stackTemplateId"
//	@Success		200				{object}	domain.GetStackTemplateResponse
//	@Router			/admin/1.0/stack-templates/{stackTemplateId} [get]
//	@Security		JWT
func (h *StackTemplateHandler) GetStackTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackTemplateId"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	stackTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	stackTemplate, err := h.usecase.Get(r.Context(), stackTemplateId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackTemplateResponse
	if err := serializer.Map(stackTemplate, &out.StackTemplate); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	out.StackTemplate.Organizations = make([]domain.SimpleOrganizationResponse, len(stackTemplate.Organizations))
	for i, organization := range stackTemplate.Organizations {
		if err := serializer.Map(organization, &out.StackTemplate.Organizations[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	err = json.Unmarshal(stackTemplate.Services, &out.StackTemplate.Services)
	if err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateStackTemplate godoc
//
//	@Tags			StackTemplates
//	@Summary		Update StackTemplate 'NOT IMPLEMENTED'
//	@Description	Update StackTemplate
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.UpdateStackTemplateRequest	true	"Update stack template request"
//	@Success		200		{object}	nil
//	@Router			/admin/1.0/stack-templates/{stackTemplateId} [put]
//	@Security		JWT
func (h *StackTemplateHandler) UpdateStackTemplate(w http.ResponseWriter, r *http.Request) {
	/*
		vars := mux.Vars(r)
		strId, ok := vars["stackTemplateId"]
		if !ok {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackTemplateId")))
			return
		}

		stackTemplateId, err := uuid.Parse(strId)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
			return
		}

		var dto domain.StackTemplate
		if err := serializer.Map(r, &dto); err != nil {
			log.InfoWithContext(r.Context(),err)
		}
		dto.ID = stackTemplateId

		err = h.usecase.Update(r.Context(), dto)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}
	*/

	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// DeleteStackTemplate godoc
//
//	@Tags			StackTemplates
//	@Summary		Delete StackTemplate 'NOT IMPLEMENTED'
//	@Description	Delete StackTemplate
//	@Accept			json
//	@Produce		json
//	@Param			stackTemplateId	path		string	true	"stackTemplateId"
//	@Success		200				{object}	nil
//	@Router			/admin/1.0/stack-templates/{stackTemplateId} [delete]
//	@Security		JWT
func (h *StackTemplateHandler) DeleteStackTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["stackTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackTemplateId"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// GetStackTemplateServices godoc
//
//	@Tags			StackTemplates
//	@Summary		Get GetStackTemplateServices
//	@Description	Get GetStackTemplateServices
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.GetStackTemplateServicesResponse
//	@Router			/admin/1.0/stack-templates/services [get]
//	@Security		JWT
func (h *StackTemplateHandler) GetStackTemplateServices(w http.ResponseWriter, r *http.Request) {

	var out domain.GetStackTemplateServicesResponse
	out.Services = make([]domain.StackTemplateServiceResponse, 2)
	err := json.Unmarshal([]byte(internal.SERVICE_LMA), &out.Services[0])
	if err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	err = json.Unmarshal([]byte(internal.SERVICE_SERVICE_MESH), &out.Services[1])
	if err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
