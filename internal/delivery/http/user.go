package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}
	userInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("user info not found in token")))
		return
	}

	log.Info("Send signup request to keycloak")
	ctx := r.Context()
	user := input.ToUser()
	user.Organization = domain.Organization{
		ID: userInfo.GetOrganizationId(),
	}
	user, err = u.usecase.Create(ctx, user)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}

	var out struct {
		User domain.User
	}

	out.User = *user

	ResponseJSON(w, http.StatusCreated, out)

}

func (u UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("userId not found in path")))
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}
	if user == nil {
		ResponseJSON(w, http.StatusNotFound, nil)
	}

	var out struct {
		User domain.User
	}
	out.User = *user

	ResponseJSON(w, http.StatusOK, out)
}

func (u UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := u.usecase.List(r.Context())
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
	}
	if users == nil {
		ResponseJSON(w, http.StatusNotFound, nil)
	}

	var out struct {
		Users []domain.User
	}
	for _, user := range *users {
		out.Users = append(out.Users, user)
	}

	ResponseJSON(w, http.StatusOK, out)
}

func (u UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("userId not found in path")))
		return
	}

	err := u.usecase.DeleteByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}
	ResponseJSON(w, http.StatusOK, nil)
}

func (u UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("userId not found in path")))
		return
	}

	input := domain.UpdateUserRequest{}
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
	userInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("user info not found in token")))
		return
	}

	log.Info("Send signup request to keycloak")
	ctx := r.Context()
	user := input.ToUser()
	user.Organization = domain.Organization{
		ID: userInfo.GetOrganizationId(),
	}

	originUser, err := u.usecase.GetByAccountId(ctx, userId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}
	if originUser == nil {
		user, err = u.usecase.Create(ctx, user)
		if err != nil {
			ErrorJSON(w, httpErrors.NewInternalServerError(err))
			return
		}
	}

	user, err = u.usecase.UpdateByAccountId(ctx, userId, user)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}

	var out struct {
		User domain.User
	}

	out.User = *user

	ResponseJSON(w, http.StatusOK, out)
}

func (u UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("userId not found in path")))
		return
	}

	input := domain.UpdatePasswordRequest{}
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

	err = u.usecase.UpdatePasswordByAccountId(r.Context(), userId, input.Password)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

func (u UserHandler) CheckId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("userId not found in path")))
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}

	if user != nil {
		ErrorJSON(w, httpErrors.NewConflictError(fmt.Errorf("user already exists")))
	}

	ResponseJSON(w, http.StatusNotFound, nil)
}

func NewUserHandler(h usecase.IUserUsecase) IUserHandler {
	return &UserHandler{
		usecase: h,
	}
}
