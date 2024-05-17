package http

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type AuditHandler struct {
	usecase usecase.IAuditUsecase
}

func NewAuditHandler(h usecase.Usecase) *AuditHandler {
	return &AuditHandler{
		usecase: h.Audit,
	}
}

// CreateAudit godoc
//
//	@Tags			Audits
//	@Summary		Create Audit
//	@Description	Create Audit
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateAuditRequest	true	"create audit request"
//	@Success		200		{object}	domain.CreateAuditResponse
//	@Router			/admin/audits [post]
//	@Security		JWT
func (h *AuditHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// GetAudit godoc
//
//	@Tags			Audits
//	@Summary		Get Audits
//	@Description	Get Audits
//	@Accept			json
//	@Produce		json
//	@Param			pageSize	query		string		false	"pageSize"
//	@Param			pageNumber	query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filter		query		[]string	false	"filters"
//	@Param			or			query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetAuditsResponse
//	@Router			/admin/audits [get]
//	@Security		JWT
func (h *AuditHandler) GetAudits(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	audits, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAuditsResponse
	out.Audits = make([]domain.AuditResponse, len(audits))
	for i, audit := range audits {
		if err := serializer.Map(r.Context(), audit, &out.Audits[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAudit godoc
//
//	@Tags			Audits
//	@Summary		Get Audit
//	@Description	Get Audit
//	@Accept			json
//	@Produce		json
//	@Param			auditId	path		string	true	"auditId"
//	@Success		200		{object}	domain.GetAuditResponse
//	@Router			/admin/audits/{auditId} [get]
//	@Security		JWT
func (h *AuditHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["auditId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid auditId"), "C_INVALID_AUDIT_ID", ""))
		return
	}

	auditId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_AUDIT_ID", ""))
		return
	}

	audit, err := h.usecase.Get(r.Context(), auditId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	log.Info(r.Context(), audit)

	var out domain.GetAuditResponse
	if err := serializer.Map(r.Context(), audit, &out.Audit); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)

}

// DeleteAudit godoc
//
//	@Tags			Audits
//	@Summary		Delete Audit 'NOT IMPLEMENTED'
//	@Description	Delete Audit
//	@Accept			json
//	@Produce		json
//	@Param			auditId	path		string	true	"auditId"
//	@Success		200		{object}	nil
//	@Router			/admin/audits/{auditId} [delete]
//	@Security		JWT
func (h *AuditHandler) DeleteAudit(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}
