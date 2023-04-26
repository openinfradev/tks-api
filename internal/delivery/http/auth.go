package http

import (
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	FindId(w http.ResponseWriter, r *http.Request)
	FindPassword(w http.ResponseWriter, r *http.Request)
	VerifyIdentityForLostId(w http.ResponseWriter, r *http.Request)
	VerifyIdentityForLostPassword(w http.ResponseWriter, r *http.Request)

	//Authenticate(next http.Handler) http.Handler
}
type AuthHandler struct {
	usecase usecase.IAuthUsecase
}

func NewAuthHandler(h usecase.IAuthUsecase) IAuthHandler {
	return &AuthHandler{
		usecase: h,
	}
}

// Login godoc
// @Tags Auth
// @Summary login
// @Description login
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "account info"
// @Success 200 {object} domain.LoginResponse "user detail"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	input := domain.LoginRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	user, err := h.usecase.Login(input.AccountId, input.Password, input.OrganizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var out domain.LoginResponse
	if err = domain.Map(user, &out.User); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// Logout godoc
// @Tags Auth
// @Summary logout
// @Description logout
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Router /auth/logout [post]
// @Security     JWT
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Do nothing
	// Token is not able to be expired manually. Therefore, nothing to do currently.z
	ctx := r.Context()

	sessionId, ok := request.SessionFrom(ctx)
	if !ok {
		log.Errorf("session id is not found")
		ErrorJSON(w, httpErrors.NewInternalServerError(httpErrors.InternalServerError))
		return
	}
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		log.Errorf("user info is not found")
		ErrorJSON(w, httpErrors.NewInternalServerError(httpErrors.InternalServerError))
		return
	}
	organizationId := userInfo.GetOrganizationId()

	err := h.usecase.Logout(sessionId, organizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

// FindId godoc
// @Tags Auth
// @Summary Request to find forgotten ID
// @Description This API allows users to find their account ID by submitting required information
// @Accept json
// @Produce json
// @Param body body domain.FindIdRequest true "Request body for finding the account ID including {organization ID, email, username, 6 digit code}"
// @Success 200 {object} domain.FindIdResponse
// @Failure 400 {object} httpErrors.RestError
// @Router /auth/find-id/verification [post]
func (h *AuthHandler) FindId(w http.ResponseWriter, r *http.Request) {
	input := domain.FindIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	accountId, err := h.usecase.FindId(input.Code, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}
	var out domain.FindIdResponse
	out.AccountId = accountId

	ResponseJSON(w, http.StatusOK, out)
}

// FindPassword godoc
// @Tags Auth
// @Summary Request to find forgotten password
// @Description This API allows users to reset their forgotten password by submitting required information
// @Accept json
// @Produce json
// @Param body body domain.FindPasswordRequest true "Request body for finding the password including {organization ID, email, username, Account ID, 6 digit code}"
// @Success 200
// @Failure 400 {object} httpErrors.RestError
// @Router /auth/find-password/verification [post]
func (h *AuthHandler) FindPassword(w http.ResponseWriter, r *http.Request) {
	input := domain.FindPasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.FindPassword(input.Code, input.AccountId, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// VerifyIdentityForLostId godoc
// @Tags Auth
// @Summary Request to verify identity for lost id
// @Description This API allows users to verify their identity for lost id by submitting required information
// @Accept json
// @Produce json
// @Param body body domain.VerifyIdentityForLostIdRequest true "Request body for verifying identity for lost id including {organization ID, email, username}"
// @Success 200 {object} domain.VerifyIdentityForLostIdResponse
// @Failure 400 {object} httpErrors.RestError
// @Router /auth/find-id/code [post]
func (h *AuthHandler) VerifyIdentityForLostId(w http.ResponseWriter, r *http.Request) {
	input := domain.VerifyIdentityForLostIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.VerifyIdentity("", input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}
	var out domain.VerifyIdentityForLostIdResponse
	out.ValidityPeriod = fmt.Sprintf("%.0f", internal.EmailCodeExpireTime.Seconds())

	ResponseJSON(w, http.StatusOK, out)
}

// VerifyIdentityForLostPassword godoc
// @Tags Auth
// @Summary Request to verify identity for lost password
// @Description This API allows users to verify their identity for lost password by submitting required information
// @Accept json
// @Produce json
// @Param body body domain.VerifyIdentityForLostPasswordRequest true "Request body for verifying identity for lost password including {organization ID, email, username, Account ID}"
// @Success 200 {object} domain.VerifyIdentityForLostPasswordResponse
// @Failure 400 {object} httpErrors.RestError
// @Router /auth/find-password/code [post]
func (h *AuthHandler) VerifyIdentityForLostPassword(w http.ResponseWriter, r *http.Request) {
	input := domain.VerifyIdentityForLostPasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = h.usecase.VerifyIdentity(input.AccountId, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}
	var out domain.VerifyIdentityForLostPasswordResponse
	out.ValidityPeriod = fmt.Sprintf("%.0f", internal.EmailCodeExpireTime.Seconds())

	ResponseJSON(w, http.StatusOK, out)
}
