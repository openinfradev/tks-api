package http

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type CloudAccountHandler struct {
	usecase usecase.ICloudAccountUsecase
}

func NewCloudAccountHandler(h usecase.Usecase) *CloudAccountHandler {
	return &CloudAccountHandler{
		usecase: h.CloudAccount,
	}
}

// CreateCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Create CloudAccount
//	@Description	Create CloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"organizationId"
//	@Param			body			body		domain.CreateCloudAccountRequest	true	"create cloud setting request"
//	@Success		200				{object}	domain.CreateCloudAccountResponse
//	@Router			/organizations/{organizationId}/cloud-accounts [post]
//	@Security		JWT
func (h *CloudAccountHandler) CreateCloudAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.CreateCloudAccountRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.CloudAccount
	if err = serializer.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}
	dto.OrganizationId = organizationId

	cloudAccountId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateCloudAccountResponse
	out.ID = cloudAccountId.String()

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Get CloudAccounts
//	@Description	Get CloudAccounts
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"organizationId"
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.GetCloudAccountsResponse
//	@Router			/organizations/{organizationId}/cloud-accounts [get]
//	@Security		JWT
func (h *CloudAccountHandler) GetCloudAccounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	cloudAccounts, err := h.usecase.Fetch(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetCloudAccountsResponse
	out.CloudAccounts = make([]domain.CloudAccountResponse, len(cloudAccounts))
	for i, cloudAccount := range cloudAccounts {
		if err := serializer.Map(cloudAccount, &out.CloudAccounts[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Get CloudAccount
//	@Description	Get CloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			cloudAccountId	path		string	true	"cloudAccountId"
//	@Success		200				{object}	domain.GetCloudAccountResponse
//	@Router			/organizations/{organizationId}/cloud-accounts/{cloudAccountId} [get]
//	@Security		JWT
func (h *CloudAccountHandler) GetCloudAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["cloudAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudAccountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	cloudAccountId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	cloudAccount, err := h.usecase.Get(r.Context(), cloudAccountId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetCloudAccountResponse
	if err := serializer.Map(cloudAccount, &out.CloudAccount); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Update CloudAccount
//	@Description	Update CloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"organizationId"
//	@Param			body			body		domain.UpdateCloudAccountRequest	true	"Update cloud setting request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/cloud-accounts/{cloudAccountId} [put]
//	@Security		JWT
func (h *CloudAccountHandler) UpdateCloudAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["cloudAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudAccountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	log.DebugWithContext(r.Context(), "[TODO] organization check", organizationId)

	cloudAccountId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	input := domain.UpdateCloudAccountRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.CloudAccount
	if err = serializer.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}
	dto.ID = cloudAccountId
	dto.OrganizationId = organizationId

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Delete CloudAccount
//	@Description	Delete CloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"organizationId"
//	@Param			body			body		domain.DeleteCloudAccountRequest	true	"Delete cloud setting request"
//	@Param			cloudAccountId	path		string								true	"cloudAccountId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/cloud-accounts/{cloudAccountId} [delete]
//	@Security		JWT
func (h *CloudAccountHandler) DeleteCloudAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudAccountId, ok := vars["cloudAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudAccountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	parsedId, err := uuid.Parse(cloudAccountId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	input := domain.DeleteCloudAccountRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.CloudAccount
	if err = serializer.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}
	dto.ID = parsedId

	err = h.usecase.Delete(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteForceCloudAccount godoc
//
//	@Tags			CloudAccounts
//	@Summary		Delete Force CloudAccount
//	@Description	Delete Force CloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			cloudAccountId	path		string	true	"cloudAccountId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/error [delete]
//	@Security		JWT
func (h *CloudAccountHandler) DeleteForceCloudAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudAccountId, ok := vars["cloudAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudAccountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	parsedId, err := uuid.Parse(cloudAccountId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	err = h.usecase.DeleteForce(r.Context(), parsedId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// CheckCloudAccountName godoc
//
//	@Tags			CloudAccounts
//	@Summary		Check name for cloudAccount
//	@Description	Check name for cloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			name			path		string	true	"name"
//	@Success		200				{object}	domain.CheckCloudAccountNameResponse
//	@Router			/organizations/{organizationId}/cloud-accounts/name/{name}/existence [GET]
//	@Security		JWT
func (h *CloudAccountHandler) CheckCloudAccountName(w http.ResponseWriter, r *http.Request) {
	user, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", ""))
		return
	}

	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name"), "CA_INVALID_CLOUD_ACCOUNT_NAME", ""))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(r.Context(), user.GetOrganizationId(), name)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckCloudAccountNameResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// CheckAwsAccountId godoc
//
//	@Tags			CloudAccounts
//	@Summary		Check awsAccountId for cloudAccount
//	@Description	Check awsAccountId for cloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			awsAccountId	path		string	true	"awsAccountId"
//	@Success		200				{object}	domain.CheckCloudAccountAwsAccountIdResponse
//	@Router			/organizations/{organizationId}/cloud-accounts/aws-account-id/{awsAccountId}/existence [GET]
//	@Security		JWT
func (h *CloudAccountHandler) CheckAwsAccountId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	awsAccountId, ok := vars["awsAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid accountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	exist := true
	_, err := h.usecase.GetByAwsAccountId(r.Context(), awsAccountId)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckCloudAccountAwsAccountIdResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetResourceQuota godoc
//
//	@Tags			CloudAccounts
//	@Summary		Get resource quota by cloudAccount
//	@Description	Get resource quota by cloudAccount
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			cloudAccountId	path		string	true	"cloudAccountId"
//	@Success		200				{object}	domain.GetCloudAccountResourceQuotaResponse
//	@Router			/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/quota [GET]
//	@Security		JWT
func (h *CloudAccountHandler) GetResourceQuota(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["cloudAccountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudAccountId"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	cloudAccountId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_CLOUD_ACCOUNT_ID", ""))
		return
	}

	available, resourceQuota, err := h.usecase.GetResourceQuota(r.Context(), cloudAccountId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetCloudAccountResourceQuotaResponse
	if err := serializer.Map(resourceQuota, &out.ResourceQuota); err != nil {
		log.InfoWithContext(r.Context(), err)
	}
	out.Available = available

	ResponseJSON(w, r, http.StatusOK, out)
}
