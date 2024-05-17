package http

import (
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/middleware/audit"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/serializer"
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

	VerifyToken(w http.ResponseWriter, r *http.Request)
	//Authenticate(next http.Handler) http.Handler
}
type AuthHandler struct {
	usecase        usecase.IAuthUsecase
	auditUsecase   usecase.IAuditUsecase
	projectUsecase usecase.IProjectUsecase
}

func NewAuthHandler(h usecase.Usecase) IAuthHandler {
	return &AuthHandler{
		usecase:        h.Auth,
		auditUsecase:   h.Audit,
		projectUsecase: h.Project,
	}
}

// Login godoc
//
//	@Tags			Auth
//	@Summary		login
//	@Description	login
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.LoginRequest		true	"account info"
//	@Success		200		{object}	domain.LoginResponse	"user detail"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	input := domain.LoginRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	user, err := h.usecase.Login(r.Context(), input.AccountId, input.Password, input.OrganizationId)
	if err != nil {
		errorResponse, _ := httpErrors.ErrorResponse(err)
		_, _ = h.auditUsecase.Create(r.Context(), model.Audit{
			OrganizationId: input.OrganizationId,
			Group:          "Auth",
			Message:        fmt.Sprintf("[%s]님이 로그인에 실패하였습니다.", input.AccountId),
			Description:    errorResponse.Text(),
			ClientIP:       audit.GetClientIpAddress(w, r),
			UserId:         nil,
			UserAccountId:  input.AccountId,
		})
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	} else {
		_, _ = h.auditUsecase.Create(r.Context(), model.Audit{
			OrganizationId: input.OrganizationId,
			Group:          "Auth",
			Message:        fmt.Sprintf("[%s]님이 로그인 하였습니다.", input.AccountId),
			Description:    "",
			ClientIP:       audit.GetClientIpAddress(w, r),
			UserId:         &user.ID,
		})
	}

	var cookies []*http.Cookie
	if targetCookies, err := h.usecase.SingleSignIn(r.Context(), input.OrganizationId, input.AccountId, input.Password); err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
	} else {
		cookies = append(cookies, targetCookies...)
	}

	if len(cookies) > 0 {
		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
		}
	}

	var out domain.LoginResponse
	if err = serializer.Map(r.Context(), user, &out.User); err != nil {
		log.Error(r.Context(), err)
	}
	for _, role := range user.Roles {
		out.User.Roles = append(out.User.Roles, domain.SimpleRoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// Logout godoc
//
//	@Tags			Auth
//	@Summary		logout
//	@Description	logout
//	@Accept			json
//	@Produce		json
//	@Router			/auth/logout [post]
//	@Security		JWT
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionId, ok := request.SessionFrom(ctx)
	if !ok {
		log.Errorf(r.Context(), "session id is not found")
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("session id is not found"), "A_NO_SESSION", ""))
		return
	}
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		log.Errorf(r.Context(), "user info is not found")
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user info is not found"), "A_NO_SESSION", ""))
		return
	}
	organizationId := userInfo.GetOrganizationId()

	err := h.usecase.Logout(r.Context(), sessionId, organizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	var cookies []*http.Cookie
	redirectUrl, targetCookies, err := h.usecase.SingleSignOut(r.Context(), organizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
	}
	cookies = append(cookies, targetCookies...)
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
		}
	}

	//추후 사용을 위해 주석 처리
	//http.Redirect(w, r, redirectUrl, http.StatusFound)
	//ResponseJSON(w, r, http.StatusFound, nil)

	//추후 사용을 위한 임시 코드
	_ = redirectUrl

	ResponseJSON(w, r, http.StatusOK, nil)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

// FindId godoc
//
//	@Tags			Auth
//	@Summary		Request to find forgotten ID
//	@Description	This API allows users to find their account ID by submitting required information
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.FindIdRequest	true	"Request body for finding the account ID including {organization ID, email, username, 6 digit code}"
//	@Success		200		{object}	domain.FindIdResponse
//	@Failure		400		{object}	httpErrors.RestError
//	@Router			/auth/find-id/verification [post]
func (h *AuthHandler) FindId(w http.ResponseWriter, r *http.Request) {
	input := domain.FindIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	accountId, err := h.usecase.FindId(r.Context(), input.Code, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}
	var out domain.FindIdResponse
	out.AccountId = accountId

	ResponseJSON(w, r, http.StatusOK, out)
}

// FindPassword godoc
//
//	@Tags			Auth
//	@Summary		Request to find forgotten password
//	@Description	This API allows users to reset their forgotten password by submitting required information
//	@Accept			json
//	@Produce		json
//	@Param			body	body	domain.FindPasswordRequest	true	"Request body for finding the password including {organization ID, email, username, Account ID, 6 digit code}"
//	@Success		200
//	@Failure		400	{object}	httpErrors.RestError
//	@Router			/auth/find-password/verification [post]
func (h *AuthHandler) FindPassword(w http.ResponseWriter, r *http.Request) {
	input := domain.FindPasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.FindPassword(r.Context(), input.Code, input.AccountId, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// VerifyIdentityForLostId godoc
//
//	@Tags			Auth
//	@Summary		Request to verify identity for lost id
//	@Description	This API allows users to verify their identity for lost id by submitting required information
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.VerifyIdentityForLostIdRequest	true	"Request body for verifying identity for lost id including {organization ID, email, username}"
//	@Success		200		{object}	domain.VerifyIdentityForLostIdResponse
//	@Failure		400		{object}	httpErrors.RestError
//	@Router			/auth/find-id/code [post]
func (h *AuthHandler) VerifyIdentityForLostId(w http.ResponseWriter, r *http.Request) {
	input := domain.VerifyIdentityForLostIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.VerifyIdentity(r.Context(), "", input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	var out domain.VerifyIdentityForLostIdResponse
	out.ValidityPeriod = fmt.Sprintf("%.0f", internal.EmailCodeExpireTime.Seconds())

	ResponseJSON(w, r, http.StatusOK, out)
}

// VerifyIdentityForLostPassword godoc
//
//	@Tags			Auth
//	@Summary		Request to verify identity for lost password
//	@Description	This API allows users to verify their identity for lost password by submitting required information
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.VerifyIdentityForLostPasswordRequest	true	"Request body for verifying identity for lost password including {organization ID, email, username, Account ID}"
//	@Success		200		{object}	domain.VerifyIdentityForLostPasswordResponse
//	@Failure		400		{object}	httpErrors.RestError
//	@Router			/auth/find-password/code [post]
func (h *AuthHandler) VerifyIdentityForLostPassword(w http.ResponseWriter, r *http.Request) {
	input := domain.VerifyIdentityForLostPasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.VerifyIdentity(r.Context(), input.AccountId, input.Email, input.UserName, input.OrganizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	var out domain.VerifyIdentityForLostPasswordResponse
	out.ValidityPeriod = fmt.Sprintf("%.0f", internal.EmailCodeExpireTime.Seconds())

	ResponseJSON(w, r, http.StatusOK, out)
}

// VerifyToken godoc
//	@Tags			Auth
//	@Summary		verify token
//	@Description	verify token
//	@Success		200	{object}	nil
//	@Failure		401	{object}	nil
//	@Router			/auth/verify-token [get]

func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	token, ok := request.TokenFrom(r.Context())
	if !ok {
		log.Errorf(r.Context(), "token is not found")
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("token is not found"), "C_INTERNAL_ERROR", ""))
		return
	}

	isActive, err := h.usecase.VerifyToken(r.Context(), token)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	if !isActive {
		ErrorJSON(w, r, httpErrors.NewUnauthorizedError(fmt.Errorf("token is not active"), "A_EXPIRED_TOKEN", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
