package http

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"net/http"

	"github.com/openinfradev/tks-api/pkg/log"

	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	FindId(w http.ResponseWriter, r *http.Request)
	FindPassword(w http.ResponseWriter, r *http.Request)

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
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Do nothing
	// Token is not able to be expired manually. Therefore, nothing to do currently.z
	ctx := r.Context()
	accessToken, ok := request.TokenFrom(ctx)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("token not found")))
		return
	}
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("user not found")))
		return
	}
	err := h.usecase.Logout(accessToken, userInfo.GetUserId().String(), userInfo.GetOrganizationId())
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

func (h *AuthHandler) FindId(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (h *AuthHandler) FindPassword(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}
