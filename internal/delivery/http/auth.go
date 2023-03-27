package http

import (
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"

	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
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
// @Success 200 {object} domain.User "user detail"
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
		log.Error("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}

	var out struct {
		User domain.User `json:"user"`
	}

	out.User = user

	//_ = h.Repository.AddHistory(user.ID.String(), "", "login", fmt.Sprintf("[%s] 님이 로그인하였습니다.", input.AccountId))

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
	// Token is not able to be expired manually. Therefore, nothing to do currently.
	ResponseJSON(w, http.StatusOK, nil)
}

func (h *AuthHandler) FindId(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *AuthHandler) FindPassword(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}
