package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type SystemNotificationHandler struct {
	usecase usecase.ISystemNotificationUsecase
}

func NewSystemNotificationHandler(h usecase.Usecase) *SystemNotificationHandler {
	return &SystemNotificationHandler{
		usecase: h.SystemNotification,
	}
}

// CreateSystemNotification godoc
//
//	@Tags			SystemNotifications
//	@Summary		Create systemNotification. ADMIN ONLY
//	@Description	Create systemNotification. ADMIN ONLY
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	nil
//	@Router			/system-api/organizations/{organizationId}/system-notifications [post]
//	@Security		JWT
func (h *SystemNotificationHandler) CreateSystemNotification(w http.ResponseWriter, r *http.Request) {

	/*
		INFO[2023-04-26 18:14:11] body : {"receiver":"webhook-systemNotification","status":"firing","systemNotifications":[{"status":"firing","labels":{"systemNotificationname":"TestSystemNotification1"},"annotations":{},"startsAt":"2023-04-26T09:14:01.489894015Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"0dafe30dffce9487"}],"groupLabels":{"systemNotificationname":"TestSystemNotification1"},"commonLabels":{"systemNotificationname":"TestSystemNotification1"},"commonAnnotations":{},"externalURL":"http://lma-systemNotificationmanager.lma:9093","version":"4","groupKey":"{}:{systemNotificationname=\"TestSystemNotification1\"}","truncatedSystemNotifications":0}
		INFO[2023-04-26 18:14:11] {"receiver":"webhook-systemNotification","status":"firing","systemNotifications":[{"status":"firing","labels":{"systemNotificationname":"TestSystemNotification1"},"annotations":{},"startsAt":"2023-04-26T09:14:01.489894015Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"0dafe30dffce9487"}],"groupLabels":{"systemNotificationname":"TestSystemNotification1"},"commonLabels":{"systemNotificationname":"TestSystemNotification1"},"commonAnnotations":{},"externalURL":"http://lma-systemNotificationmanager.lma:9093","version":"4","groupKey":"{}:{systemNotificationname=\"TestSystemNotification1\"}","truncatedSystemNotifications":0}
	*/

	/*
		// webhook 으로 부터 받은 body parse
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(r.Context(),err)
		}
		bodyString := string(bodyBytes)
		log.Info(r.Context(),bodyString)
	*/

	// 외부로부터(systemNotification manager) 오는 데이터이므로, dto 변환없이 by-pass 처리한다.
	input := domain.CreateSystemNotificationRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.Create(r.Context(), input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetSystemNotification godoc
//
//	@Tags			SystemNotifications
//	@Summary		Get SystemNotifications
//	@Description	Get SystemNotifications
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"organizationId"
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.GetSystemNotificationsResponse
//	@Router			/organizations/{organizationId}/system-notifications [get]
//	@Security		JWT
func (h *SystemNotificationHandler) GetSystemNotifications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	// convert status
	for i, filter := range pg.GetFilters() {
		if filter.Column == "status" {
			for j, value := range filter.Values {
				switch value {
				case "CREATED":
					pg.GetFilters()[i].Values[j] = "0"
				case "INPROGRESS":
					pg.GetFilters()[i].Values[j] = "1"
				case "CLOSED":
					pg.GetFilters()[i].Values[j] = "2"
				case "ERROR":
					pg.GetFilters()[i].Values[j] = "3"
				}
			}
		}
	}

	systemNotifications, err := h.usecase.Fetch(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationsResponse
	out.SystemNotifications = make([]domain.SystemNotificationResponse, len(systemNotifications))
	for i, systemNotification := range systemNotifications {
		if err := serializer.Map(r.Context(), systemNotification, &out.SystemNotifications[i]); err != nil {
			log.Info(r.Context(), err)
		}

		outSystemNotificationActions := make([]domain.SystemNotificationActionResponse, len(systemNotification.SystemNotificationActions))
		for j, systemNotificationAction := range systemNotification.SystemNotificationActions {
			if err := serializer.Map(r.Context(), systemNotificationAction, &outSystemNotificationActions[j]); err != nil {
				log.Info(r.Context(), err)
			}
		}
		out.SystemNotifications[i].SystemNotificationActions = outSystemNotificationActions
		if len(outSystemNotificationActions) > 0 {
			out.SystemNotifications[i].LastTaker = outSystemNotificationActions[0].Taker
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetSystemNotification godoc
//
//	@Tags			SystemNotifications
//	@Summary		Get SystemNotification
//	@Description	Get SystemNotification
//	@Accept			json
//	@Produce		json
//	@Param			organizationId			path		string	true	"organizationId"
//	@Param			systemNotificationId	path		string	true	"systemNotificationId"
//	@Success		200						{object}	domain.GetSystemNotificationResponse
//	@Router			/organizations/{organizationId}/system-notifications/{systemNotificationId} [get]
//	@Security		JWT
func (h *SystemNotificationHandler) GetSystemNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid systemNotificationId"), "", ""))
		return
	}

	systemNotificationId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "", ""))
		return
	}

	systemNotification, err := h.usecase.Get(r.Context(), systemNotificationId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationResponse
	if err := serializer.Map(r.Context(), systemNotification, &out.SystemNotification); err != nil {
		log.Info(r.Context(), err)
	}
	outSystemNotificationActions := make([]domain.SystemNotificationActionResponse, len(systemNotification.SystemNotificationActions))
	for j, systemNotificationAction := range systemNotification.SystemNotificationActions {
		if err := serializer.Map(r.Context(), systemNotificationAction, &outSystemNotificationActions[j]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}
	out.SystemNotification.SystemNotificationActions = outSystemNotificationActions

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateSystemNotification godoc
//
//	@Tags			SystemNotifications
//	@Summary		Update SystemNotification
//	@Description	Update SystemNotification
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"organizationId"
//	@Param			body			body		domain.UpdateSystemNotificationRequest	true	"Update cloud setting request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/system-notifications/{systemNotificationId} [put]
//	@Security		JWT
func (h *SystemNotificationHandler) UpdateSystemNotification(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("Need implementation"))
}

// DeleteSystemNotification godoc
//
//	@Tags			SystemNotifications
//	@Summary		Delete SystemNotification
//	@Description	Delete SystemNotification
//	@Accept			json
//	@Produce		json
//	@Param			organizationId			path		string	true	"organizationId"
//	@Param			systemNotificationId	path		string	true	"systemNotificationId"
//	@Success		200						{object}	nil
//	@Router			/organizations/{organizationId}/system-notifications/{systemNotificationId} [delete]
//	@Security		JWT
func (h *SystemNotificationHandler) DeleteSystemNotification(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, r, fmt.Errorf("Need implementation"))
}

func (h *SystemNotificationHandler) SystemNotificationTest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	log.Info(r.Context(), "TEST ", body)
}

// CreateSystemNotificationAction godoc
//
//	@Tags			SystemNotifications
//	@Summary		Create systemNotification action
//	@Description	Create systemNotification action
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/system-notifications/{systemNotificationId}/actions [post]
//	@Security		JWT
func (h *SystemNotificationHandler) CreateSystemNotificationAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid systemNotificationId"), "", ""))
		return
	}

	systemNotificationId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "", ""))
		return
	}

	input := domain.CreateSystemNotificationActionRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	log.Info(r.Context(), "systemNotification : ", helper.ModelToJson(input))

	var dto model.SystemNotificationAction
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.SystemNotificationId = systemNotificationId

	systemNotificationAction, err := h.usecase.CreateSystemNotificationAction(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateSystemNotificationActionResponse
	out.ID = systemNotificationAction.String()
	ResponseJSON(w, r, http.StatusOK, out)
}
