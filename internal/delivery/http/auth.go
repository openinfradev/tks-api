package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	user, err := h.usecase.Login(input.AccountId, input.Password, input.OrganizationId)
	if err != nil {
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

}

// CreateUser godoc
// @Tags Users
// @Summary CreateUser
// @Description CreateUser
// @Accept json
// @Produce json
// @Param body body domain.CreateUserRequest true "user info"
// @Success 200 {object} domain.User
// @Router /users [post]
// @Security     JWT
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateUserRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	if err = json.Unmarshal(body, &input); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	token, ok := request.TokenFrom(r.Context())
	if !ok {
		ErrorJSON(w, errors.New("token not found"))
		return
	}
	log.Info("Send signup request to keycloak")
	user, err := h.usecase.Register(input.AccountId, input.Password, input.Name, input.OrganizationName, input.Role, token)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = user

	ResponseJSON(w, http.StatusOK, out)
}

// GetUsers godoc
// @Tags Users
// @Summary GetUsers
// @Description GetUsers
// @Accept json
// @Produce json
// @Param body body domain.CreateUserRequest true "user info"
// @Success 200 {object} domain.User
// @Router /users [post]
// @Security     JWT
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, fmt.Errorf("Need implementation"))
}

// GetUser godoc
// @Tags Users
// @Summary GetUser
// @Description GetUser
// @Accept json
// @Produce json
// @Param userId path string true "user uid"
// @Success 200 {object} domain.User
// @Router /users/:userId [get]
// @Security     JWT
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ErrorJSON(w, fmt.Errorf("Need implementation"))
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
		ErrorJSON(w, err)
		return
	}

	var out struct {
		Roles []domain.Role
	}
	out.Roles = roles

	ResponseJSON(w, http.StatusOK, out)
}
