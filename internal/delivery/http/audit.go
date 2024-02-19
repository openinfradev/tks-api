package http

import (
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal/usecase"
)

type AuditHandler struct {
	usecase usecase.IAuditUsecase
}

func NewAuditHandler(h usecase.IAuditUsecase) *AuditHandler {
	return &AuditHandler{
		usecase: h,
	}
}

// CreateAudit godoc
// @Tags        Audits
// @Summary     Create Audit
// @Description Create Audit
// @Accept      json
// @Produce     json
// @Param       body body     domain.CreateAuditRequest true "create audit request"
// @Success     200  {object} domain.CreateAuditResponse
// @Router      /audits [post]
// @Security    JWT
func (h *AuditHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// GetAudit godoc
// @Tags        Audits
// @Summary     Get Audits
// @Description Get Audits
// @Accept      json
// @Produce     json
// @Param       limit       query    string   false "pageSize"
// @Param       page        query    string   false "pageNumber"
// @Param       soertColumn query    string   false "sortColumn"
// @Param       sortOrder   query    string   false "sortOrder"
// @Param       filter     	query    []string false "filters"
// @Param       or     		query    []string false "filters"
// @Success     200         {object} domain.GetAuditsResponse
// @Router      /audits [get]
// @Security    JWT
func (h *AuditHandler) GetAudits(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// GetAudit godoc
// @Tags        Audits
// @Summary     Get Audit
// @Description Get Audit
// @Accept      json
// @Produce     json
// @Param       auditId path     string true "auditId"
// @Success     200             {object} domain.GetAuditResponse
// @Router      /audits/{auditId} [get]
// @Security    JWT
func (h *AuditHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// DeleteAudit godoc
// @Tags        Audits
// @Summary     Delete Audit 'NOT IMPLEMENTED'
// @Description Delete Audit
// @Accept      json
// @Produce     json
// @Param       auditId path     string true "auditId"
// @Success     200             {object} nil
// @Router      /audits/{auditId} [delete]
// @Security    JWT
func (h *AuditHandler) DeleteAudit(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}
