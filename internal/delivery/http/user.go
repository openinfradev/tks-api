package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/usecase"
)

type UserHandler struct {
	usecase usecase.IUserUsecase
}

func NewUserHandler(h usecase.IUserUsecase) *UserHandler {
	return &UserHandler{
		usecase: h,
	}
}

// Signin godoc
// @Tags Users
// @Summary signin
// @Description signin
// @Accept json
// @Produce json
// @Success 200 {object} domain.User
// @Router /signin [post]
func (h *UserHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
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

	user, err := h.usecase.GetByAccountId(input.AccountId, input.Password)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Not found user. %s", err), http.StatusBadRequest)
		return
	}

	if !helper.CheckPasswordHash(user.Password, input.Password) {
		ErrorJSON(w, "Invalid accountId and password ", http.StatusBadRequest)
		return
	}

	accessToken, err := helper.CreateJWT(user.AccountId, user.Password)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("failed to create token. err : %s", err), http.StatusBadRequest)
		return
	}

	var out struct {
		User domain.User `json:"user"`
	}

	user.Token = accessToken

	out.User = user

	//_ = h.Repository.AddHistory(user.Id.String(), "", "signin", fmt.Sprintf("[%s] 님이 로그인하였습니다.", input.AccountId))

	ResponseJSON(w, out, http.StatusOK)

}
