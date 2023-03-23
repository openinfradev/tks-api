package http

import (
	"encoding/json"
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

type CloudSettingHandler struct {
	usecase usecase.ICloudSettingUsecase
}

func NewCloudSettingHandler(h usecase.ICloudSettingUsecase) *CloudSettingHandler {
	return &CloudSettingHandler{
		usecase: h,
	}
}

// CreateCloudSetting godoc
// @Tags CloudSettings
// @Summary Create CloudSetting
// @Description Create CloudSetting
// @Accept json
// @Produce json
// @Param body body domain.CreateCloudSettingRequest true "create cloud setting request"
// @Success 200 {object} object
// @Router /cloud-settings [post]
// @Security     JWT
func (h *CloudSettingHandler) CreateCloudSetting(w http.ResponseWriter, r *http.Request) {
	organizationId, userId, _ := GetSession(r)

	input := domain.CreateCloudSettingRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	resource := ""
	cloudSettingId, err := h.usecase.Create(organizationId, input, resource, userId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	log.Info("Newly created cloud setting : ", cloudSettingId)

	var out struct {
		CloudSettingId string `json:"cloudSettingId"`
	}
	out.CloudSettingId = cloudSettingId.String()

	ResponseJSON(w, http.StatusOK, out)

}

// GetCloudSetting godoc
// @Tags CloudSettings
// @Summary Get CloudSetting
// @Description Get CloudSetting
// @Accept json
// @Produce json
// @Success 200 {object} domain.CloudSetting
// @Router /cloud-settings [get]
// @Security     JWT
func (h *CloudSettingHandler) GetCloudSetting(w http.ResponseWriter, r *http.Request) {
	organizationId, _, _ := GetSession(r)

	cloudSetting, err := h.usecase.GetByOrganizationId(organizationId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out struct {
		CloudSetting domain.CloudSetting `json:"cloudSetting"`
	}

	out.CloudSetting = cloudSetting

	ResponseJSON(w, http.StatusOK, out)
}

// GetCloudSettingById godoc
// @Tags CloudSettings
// @Summary Get cloudSetting by cloudSettingId
// @Description Get cloudSetting by cloudSettingId
// @Accept json
// @Produce json
// @Success 200 {object} domain.CloudSetting
// @Router /cloud-settings/{cloudSettingId} [get]
// @Security     JWT
func (h *CloudSettingHandler) GetCloudSettingById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudSettingId, ok := vars["cloudSettingId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudSettingId")))
		return
	}

	parsedId, err := uuid.Parse(cloudSettingId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.WithMessage(err, "Failed to parse uuid")))
		return
	}

	cloudSetting, err := h.usecase.Get(parsedId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out struct {
		CloudSetting domain.CloudSetting `json:"cloudSetting"`
	}

	out.CloudSetting = cloudSetting

	ResponseJSON(w, http.StatusOK, out)
}

// DeleteCloudSetting godoc
// @Tags CloudSettings
// @Summary Delete CloudSetting
// @Description Delete CloudSetting
// @Accept json
// @Produce json
// @Param cloudSettingId path string true "cloudSettingId"
// @Success 200 {object} domain.CloudSetting
// @Router /cloud-settings/{cloudSettingId} [delete]
// @Security     JWT
func (h *CloudSettingHandler) DeleteCloudSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudSettingId, ok := vars["cloudSettingId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudSettingId")))
		return
	}

	parsedId, err := uuid.Parse(cloudSettingId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.WithMessage(err, "Failed to parse uuid")))
		return
	}

	err = h.usecase.Delete(parsedId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}
