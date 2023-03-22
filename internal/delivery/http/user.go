package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"io"
	"net/http"
)

type IUserHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdatePassword(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	CheckId(w http.ResponseWriter, r *http.Request)
}

type UserHandler struct {
	usecase usecase.IUserUsecase
}

func (u UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO implement validation

	// Parse request body
	input := domain.CreateUserRequest{}
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
	userInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, "user not found in Token", http.StatusBadRequest)
		return
	}

	log.Info("Send signup request to keycloak")
	ctx := r.Context()
	user := input.ToUser()
	user.Organization = domain.Organization{
		Name: userInfo.GetOrganization(),
	}
	user, err = u.usecase.Create(ctx, user)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = *user

	ResponseJSON(w, out, "", http.StatusCreated)

}

func (u UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, "user not found", http.StatusBadRequest)
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user == nil {
		ResponseJSON(w, nil, "", http.StatusNoContent)
	}

	var out struct {
		User domain.User
	}
	out.User = *user

	ResponseJSON(w, out, "", http.StatusOK)
}

func (u UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := u.usecase.List(r.Context())
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	}
	if users == nil {
		ResponseJSON(w, nil, "", http.StatusNoContent)
	}

	var out struct {
		Users []domain.User
	}
	for _, user := range *users {
		out.Users = append(out.Users, user)
	}

	ResponseJSON(w, out, "", http.StatusOK)
}

func (u UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, "user not found", http.StatusBadRequest)
		return
	}

	err := u.usecase.DeleteByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ResponseJSON(w, nil, "", http.StatusOK)
}

func (u UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, "user not found", http.StatusBadRequest)
		return
	}

	input := domain.UpdateUserRequest{}
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
	userInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, "user not found in Token", http.StatusBadRequest)
		return
	}

	log.Info("Send signup request to keycloak")
	ctx := r.Context()
	user := input.ToUser()
	user.Organization = domain.Organization{
		Name: userInfo.GetOrganization(),
	}

	originUser, err := u.usecase.GetByAccountId(ctx, userId)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if originUser == nil {
		user, err = u.usecase.Create(ctx, user)
		if err != nil {
			InternalServerError(w, err)
			return
		}
	}

	user, err = u.usecase.UpdateByAccountId(ctx, userId, user)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var out struct {
		User domain.User
	}

	out.User = *user

	ResponseJSON(w, out, "", http.StatusOK)

}

func (u UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, "user not found", http.StatusBadRequest)
		return
	}

	input := domain.UpdatePasswordRequest{}
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

	err = u.usecase.UpdatePasswordByAccountId(r.Context(), userId, input.Password)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ResponseJSON(w, nil, "", http.StatusOK)
}

func (u UserHandler) CheckId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, "user not found", http.StatusBadRequest)
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		ResponseJSON(w, nil, "", http.StatusConflict)
	}

	ResponseJSON(w, nil, "", http.StatusNotFound)
}

func NewUserHandler(h usecase.IUserUsecase) IUserHandler {
	return &UserHandler{
		usecase: h,
	}
}
