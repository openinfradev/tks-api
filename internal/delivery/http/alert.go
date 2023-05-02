package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type AlertHandler struct {
	usecase usecase.IAlertUsecase
}

func NewAlertHandler(h usecase.IAlertUsecase) *AlertHandler {
	return &AlertHandler{
		usecase: h,
	}
}

// CreateAlert godoc
// @Tags Alerts
// @Summary Create alert. ADMIN ONLY
// @Description Create alert. ADMIN ONLY
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} nil
// @Router /system-api/organizations/{organizationId}/alerts [post]
// @Security     JWT
func (h *AlertHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {

	/*
		INFO[2023-04-26 18:14:11] body : {"receiver":"webhook-alert","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"TestAlert1"},"annotations":{},"startsAt":"2023-04-26T09:14:01.489894015Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"0dafe30dffce9487"}],"groupLabels":{"alertname":"TestAlert1"},"commonLabels":{"alertname":"TestAlert1"},"commonAnnotations":{},"externalURL":"http://lma-alertmanager.lma:9093","version":"4","groupKey":"{}:{alertname=\"TestAlert1\"}","truncatedAlerts":0}
		INFO[2023-04-26 18:14:11] {"receiver":"webhook-alert","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"TestAlert1"},"annotations":{},"startsAt":"2023-04-26T09:14:01.489894015Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"0dafe30dffce9487"}],"groupLabels":{"alertname":"TestAlert1"},"commonLabels":{"alertname":"TestAlert1"},"commonAnnotations":{},"externalURL":"http://lma-alertmanager.lma:9093","version":"4","groupKey":"{}:{alertname=\"TestAlert1\"}","truncatedAlerts":0}
	*/

	/*
		// webhook 으로 부터 받은 body parse
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
		}
		bodyString := string(bodyBytes)
		log.Info(bodyString)
	*/

	// 외부로부터(alert manager) 오는 데이터이므로, dto 변환없이 by-pass 처리한다.
	input := domain.CreateAlertRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.Create(r.Context(), input)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// GetAlert godoc
// @Tags Alerts
// @Summary Get Alerts
// @Description Get Alerts
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.GetAlertsResponse
// @Router /organizations/{organizationId}/alerts [get]
// @Security     JWT
func (h *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	alerts, err := h.usecase.Fetch(organizationId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetAlertsResponse
	out.Alerts = make([]domain.AlertResponse, len(alerts))
	for i, alert := range alerts {
		if err := domain.Map(alert, &out.Alerts[i]); err != nil {
			log.Info(err)
		}

		//out.Alerts[i].RawData = fmt.Sprintf("%s", alert.RawData)
		log.Info(alert.FiredAt)

		outAlertActions := make([]domain.AlertActionResponse, len(alert.AlertActions))
		for j, alertAction := range alert.AlertActions {
			if err := domain.Map(alertAction, &outAlertActions[j]); err != nil {
				log.Info(err)
			}
		}
		out.Alerts[i].AlertActions = outAlertActions
		if len(outAlertActions) > 0 {
			out.Alerts[i].LastTaker = outAlertActions[0].Taker
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetAlert godoc
// @Tags Alerts
// @Summary Get Alert
// @Description Get Alert
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param alertId path string true "alertId"
// @Success 200 {object} domain.GetAlertResponse
// @Router /organizations/{organizationId}/alerts/{alertId} [get]
// @Security     JWT
func (h *AlertHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["alertId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid alertId")))
		return
	}

	alertId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	alert, err := h.usecase.Get(alertId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetAlertResponse
	if err := domain.Map(alert, &out.Alert); err != nil {
		log.Info(err)
	}
	outAlertActions := make([]domain.AlertActionResponse, len(alert.AlertActions))
	for j, alertAction := range alert.AlertActions {
		if err := domain.Map(alertAction, &outAlertActions[j]); err != nil {
			log.Info(err)
			continue
		}
	}
	out.Alert.AlertActions = outAlertActions

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateAlert godoc
// @Tags Alerts
// @Summary Update Alert
// @Description Update Alert
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param body body domain.UpdateAlertRequest true "Update cloud setting request"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/alerts/{alertId} [put]
// @Security     JWT
func (h *AlertHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, fmt.Errorf("Need implementation"))
}

// DeleteAlert godoc
// @Tags Alerts
// @Summary Delete Alert
// @Description Delete Alert
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param alertId path string true "alertId"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/alerts/{alertId} [delete]
// @Security     JWT
func (h *AlertHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, fmt.Errorf("Need implementation"))
}

func (h *AlertHandler) AlertTest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	log.Info("TEST ", body)
}

// CreateAlertAction godoc
// @Tags Alerts
// @Summary Create alert action
// @Description Create alert action
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} nil
// @Router /organizations/{organizationId}/alerts/{alertId}/actions [post]
// @Security     JWT
func (h *AlertHandler) CreateAlertAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["alertId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid alertId")))
		return
	}

	alertId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	input := domain.CreateAlertActionRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.AlertAction
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}
	dto.AlertId = alertId

	alertAction, err := h.usecase.CreateAlertAction(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateAlertActionResponse
	out.ID = alertAction.String()
	ResponseJSON(w, http.StatusOK, out)
}
