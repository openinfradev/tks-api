package http

import (
	"encoding/json"
	"fmt"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"io"
	"net/http"
)

type IAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	FindId(w http.ResponseWriter, r *http.Request)
	FindPassword(w http.ResponseWriter, r *http.Request)

	//SingUp(w http.ResponseWriter, r *http.Request)
	//GetRole(w http.ResponseWriter, r *http.Request)
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

	user, err := h.usecase.Login(input.AccountId, input.Password, input.OrganizationName)
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	input := domain.LogoutRequest{}
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

	// Do nothing
	// Token is not able to be expired manually. Therefore, nothing to do currently.

	ResponseJSON(w, nil, "", http.StatusOK)
}

func (h *AuthHandler) FindId(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *AuthHandler) FindPassword(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

//// Signup godoc
//// @Tags Auth
//// @Summary signup
//// @Description signup
//// @Accept json
//// @Produce json
//// @Param body body domain.SignUpRequest true "account info"
//// @Success 200 {object} domain.User
//// @Router /auth/signup [post]
//// @Security     JWT
//func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
//	input := domain.SignUpRequest{}
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
//		return
//	}
//
//	err = json.Unmarshal(body, &input)
//	if err != nil {
//		ErrorJSON(w, fmt.Sprintf("Invalid request. %s", err), http.StatusBadRequest)
//		return
//	}
//
//	token, ok := request.TokenFrom(r.Context())
//	if !ok {
//		InternalServerError(w, errors.New("token not found"))
//		return
//	}
//	log.Info("Send signup request to keycloak")
//	user, err := h.usecase.Register(input.AccountId, input.Password, input.Name, input.OrganizationName, input.Role, token)
//	if err != nil {
//		InternalServerError(w, err)
//		return
//	}
//
//	var out struct {
//		User domain.User
//	}
//
//	out.User = user
//
//	ResponseJSON(w, out, "", http.StatusOK)
//}

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
