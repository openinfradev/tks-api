package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"io"
	"net/http"
)

type IAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	SingUp(w http.ResponseWriter, r *http.Request)
	GetRole(w http.ResponseWriter, r *http.Request)

	//Authenticate(next http.Handler) http.Handler
}
type AuthHandler struct {
	usecase usecase.IAuthUsecase
}

func NewAuthHandler(h usecase.IAuthUsecase) *AuthHandler {
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
		return
	}

	user, err := h.usecase.Login(input.AccountId, input.Password, input.OrganizationId)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		User domain.User `json:"user"`
	}

	out.User = user

	//_ = h.Repository.AddHistory(user.ID.String(), "", "login", fmt.Sprintf("[%s] 님이 로그인하였습니다.", input.AccountId))

	ResponseJSON(w, out, "", http.StatusOK)

}

// Signup godoc
// @Tags Auth
// @Summary signup
// @Description signup
// @Accept json
// @Produce json
// @Param body body domain.SignUpRequest true "account info"
// @Success 200 {object} domain.User
// @Router /auth/signup [post]
// @Security     JWT
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	input := domain.SignUpRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		InternalServerError(w, errors.New("token not found"))
		return
	}
	log.Info("Send signup request to keycloak")
	user, err := h.usecase.Register(input.AccountId, input.Password, input.Name, input.OrganizationId, input.Role, token)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = user

	ResponseJSON(w, out, "", http.StatusOK)
}

// GetRoles godoc
// @Tags Auth
// @Summary roles
// @Description roles
// @Accept json
// @Produce json
// @Success 200 {object} []domain.Role
// @Router /auth/roles [get]
// @Security     JWT
func (h *AuthHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.usecase.FetchRoles()
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		Roles []domain.Role
	}
	out.Roles = roles

	ResponseJSON(w, out, "", http.StatusOK)
}
