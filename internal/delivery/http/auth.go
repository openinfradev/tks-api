package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
)

type AuthHandler struct {
	usecase usecase.IAuthUsecase
}

func NewAuthHandler(h usecase.IAuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: h,
	}
}

type SignInRequest struct {
	AccountId string `json:"accountId"`
	Password  string `json:"password"`
}

// Signin godoc
// @Tags Auth
// @Summary signin
// @Description signin
// @Accept json
// @Produce json
// @Param body body SignInRequest true "account info"
// @Success 200 {object} domain.User "user detail"
// @Router /auth/signin [post]
func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	input := SignInRequest{}

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

	user, err := h.usecase.Signin(input.AccountId, input.Password)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		User domain.User `json:"user"`
	}

	out.User = user

	//_ = h.Repository.AddHistory(user.ID.String(), "", "signin", fmt.Sprintf("[%s] 님이 로그인하였습니다.", input.AccountId))

	ResponseJSON(w, out, http.StatusOK)

}

type SignUpRequest struct {
	AccountId string `json:"accountId"`
	Password  string `json:"password"`
	Name      string `json:"name"`
}

// Signup godoc
// @Tags Auth
// @Summary signup
// @Description signup
// @Accept json
// @Produce json
// @Param body body SignUpRequest true "account info"
// @Success 200 {object} domain.User
// @Router /auth/signup [post]
// @Security     JWT
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	input := SignUpRequest{}
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

	user, err := h.usecase.Register(input.AccountId, input.Password, input.Name)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = user

	ResponseJSON(w, out, http.StatusOK)
}
