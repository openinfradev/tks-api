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
// @Success 200 {object} domain.CreateCloudSettingResponse
// @Router /cloud-settings [post]
// @Security     JWT
func (h *CloudSettingHandler) CreateCloudSetting(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateCloudSettingRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.CloudSetting
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}

	cloudSettingId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateCloudSettingResponse
	out.ID = cloudSettingId.String()

	ResponseJSON(w, http.StatusOK, out)
}

// GetCloudSetting godoc
// @Tags CloudSettings
// @Summary Get CloudSettings
// @Description Get CloudSettings
// @Accept json
// @Produce json
// @Param all query string false "show all organizations"
// @Success 200 {object} domain.GetCloudSettingsResponse
// @Router /cloud-settings [get]
// @Security     JWT
func (h *CloudSettingHandler) GetCloudSettings(w http.ResponseWriter, r *http.Request) {
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

	cloudSettings, err := h.usecase.Fetch(user.GetOrganizationId())
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetCloudSettingsResponse
	out.CloudSettings = make([]domain.CloudSettingResponse, len(cloudSettings))
	for i, cloudSetting := range cloudSettings {
		if err := domain.Map(cloudSetting, &out.CloudSettings[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetCloudSetting godoc
// @Tags CloudSettings
// @Summary Get CloudSetting
// @Description Get CloudSetting
// @Accept json
// @Produce json
// @Param cloudSettingId path string true "cloudSettingId"
// @Success 200 {object} domain.GetCloudSettingResponse
// @Router /cloud-settings/{cloudSettingId} [get]
// @Security     JWT
func (h *CloudSettingHandler) GetCloudSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["cloudSettingId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudSettingId")))
		return
	}

	cloudSettingId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	cloudSetting, err := h.usecase.Get(cloudSettingId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetCloudSettingResponse
	if err := domain.Map(cloudSetting, &out.CloudSetting); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateCloudSetting godoc
// @Tags CloudSettings
// @Summary Update CloudSetting
// @Description Update CloudSetting
// @Accept json
// @Produce json
// @Param body body domain.UpdateCloudSettingRequest true "Update cloud setting request"
// @Success 200 {object} nil
// @Router /cloud-settings/{cloudSettingId} [put]
// @Security     JWT
func (h *CloudSettingHandler) UpdateCloudSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["cloudSettingId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudSettingId")))
		return
	}

	cloudSeetingId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s")))
		return
	}

	input := domain.UpdateCloudSettingRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.CloudSetting
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

// DeleteCloudSetting godoc
// @Tags CloudSettings
// @Summary Delete CloudSetting
// @Description Delete CloudSetting
// @Accept json
// @Produce json
// @Param body body domain.DeleteCloudSettingRequest true "Delete cloud setting request"
// @Param cloudSettingId path string true "cloudSettingId"
// @Success 200 {object} nil
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
		ErrorJSON(w, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid")))
		return
	}

	input := domain.DeleteCloudSettingRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.CloudSetting
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

// CheckCloudSettingName godoc
// @Tags CloudSettings
// @Summary Check name for cloudSetting
// @Description Check name for cloudSetting
// @Accept json
// @Produce json
// @Param name path string true "name"
// @Success 200 {object} nil
// @Router /cloud-settings/name/{name}/existance [GET]
// @Security     JWT
func (h *CloudSettingHandler) CheckCloudSettingName(w http.ResponseWriter, r *http.Request) {
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

	var out domain.CheckCloudSettingNameResponse
	out.Existed = exist

	ResponseJSON(w, http.StatusOK, out)
}
