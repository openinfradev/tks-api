package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/log"
)

type AuthHandler struct {
	usecase usecase.IAuthUsecase
}

func NewAuthHandler(h usecase.IAuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: h,
	}
}

// Signin godoc
// @Tags Auth
// @Summary signin
// @Description signin
// @Accept json
// @Produce json
// @Param body body SignInInput true "account info"
// @Success 200 {object} domain.User "user detail"
// @Router /auth/signin [post]
type SignInInput struct {
	AccountId string `json:"accountId"`
	Password  string `json:"password"`
}

func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	input := SignInInput{}

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
		log.Error(err)
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

// Signup godoc
// @Tags Auth
// @Summary signup
// @Description signup
// @Accept json
// @Produce json
// @Success 200 {object} domain.User
// @Router /auth/signup [post]
// @Security     JWT
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
		Name      string `json:"name"`
	}
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
		ErrorJSON(w, fmt.Sprintf("Not found user. %s", err), http.StatusBadRequest)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = user

	ResponseJSON(w, out, http.StatusOK)
}
