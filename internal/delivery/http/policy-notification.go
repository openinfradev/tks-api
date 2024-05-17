package http

import (
	"fmt"
	"net/http"
	"strconv"

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

type PolicyNotificationHandler struct {
	usecase usecase.ISystemNotificationUsecase
}

func NewPolicyNotificationHandler(h usecase.Usecase) *PolicyNotificationHandler {
	return &PolicyNotificationHandler{
		usecase: h.SystemNotification,
	}
}

// GetPolicyNotification godoc
//
//	@Tags			PolicyNotifications
//	@Summary		Get PolicyNotifications
//	@Description	Get PolicyNotifications
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"organizationId"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.GetPolicyNotificationsResponse
//	@Router			/organizations/{organizationId}/policy-notifications [get]
//	@Security		JWT
func (h *PolicyNotificationHandler) GetPolicyNotifications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	for i, filter := range pg.GetFilters() {
		if filter.Column == "status" {
			for j, value := range filter.Values {
				var s domain.SystemNotificationRuleStatus
				pg.GetFilters()[i].Values[j] = strconv.Itoa(int(s.FromString(value)))
			}
		} else if filter.Column == "message_action_proposal" {
			for j, value := range filter.Values {
				val := ""
				if value == "dryrun" {
					val = "감사"
				} else {
					val = "거부"
				}
				pg.GetFilters()[i].Values[j] = val
			}
		}
	}

	policyNotifications, err := h.usecase.FetchPolicyNotifications(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetPolicyNotificationsResponse
	out.PolicyNotifications = make([]domain.PolicyNotificationResponse, len(policyNotifications))
	for i, policyNotification := range policyNotifications {
		if err := serializer.Map(r.Context(), policyNotification, &out.PolicyNotifications[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyNotification godoc
//
//	@Tags			PolicyNotifications
//	@Summary		Get PolicyNotification
//	@Description	Get PolicyNotification
//	@Accept			json
//	@Produce		json
//	@Param			organizationId			path		string	true	"organizationId"
//	@Param			policyNotificationId	path		string	true	"policyNotificationId"
//	@Success		200						{object}	domain.GetPolicyNotificationResponse
//	@Router			/organizations/{organizationId}/policy-notifications/{policyNotificationId} [get]
//	@Security		JWT
func (h *PolicyNotificationHandler) GetPolicyNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["policyNotificationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyNotificationId"), "", ""))
		return
	}

	policyNotificationId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "", ""))
		return
	}

	policyNotification, err := h.usecase.Get(r.Context(), policyNotificationId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetPolicyNotificationResponse
	if err := serializer.Map(r.Context(), policyNotification, &out.PolicyNotification); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
