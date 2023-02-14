package http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-common/pkg/log"
)

type UserJson = struct {
	Id            string    `json:"id"`
	AccountId     string    `json:"accountId"`
	Password      string    `json:"password"`
	Name          string    `json:"name"`
	Token         string    `json:"token"`
	Authorized    bool      `json:"authorized"`
	Tutorial      bool      `json:"tutorial"`
	Role          string    `json:"role"`
	Initial       bool      `json:"initial"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	ExpireTrialAt time.Time `json:"expireTrialAt"`
}

// Signin godoc
// @Tags Auth
// @Summary Sign in
// @Description Sign in
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} UserJson
// @Router /signin [post]
func (h *APIHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	var user repository.User
	err = h.Repository.GetUserByAccountId(&user, input.AccountId)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "not found user", http.StatusBadRequest)
		return
	}

	if !helper.CheckPasswordHash(user.Password, input.Password) {
		ErrorJSON(w, "not found user", http.StatusBadRequest)
		return
	}

	accessToken, err := helper.CreateJWT(user)
	if err != nil {
		ErrorJSON(w, "failed to create token", http.StatusBadRequest)
		return
	}

	// check trial
	if time.Now().After(user.CreatedAt.AddDate(0, 1, 0)) {
		ErrorJSON(w, "expired trial", http.StatusBadRequest)
		return
	}

	var out struct {
		User UserJson `json:"user"`
	}

	reflectUser(&out.User, user)
	out.User.Token = accessToken
	out.User.ExpireTrialAt = user.CreatedAt.AddDate(0, 1, 0)

	//_ = h.Repository.AddHistory(user.Id.String(), "", "signin", fmt.Sprintf("[%s] 님이 로그인하였습니다.", input.AccountId))

	ResponseJSON(w, out, http.StatusOK)

}

// UpdatePassword godoc
// @Tags Auth
// @Summary Sign in
// @Description Sign in
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} UserJson
// @Router /users/{userId}/password [put]
func (h *APIHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		log.Error("Failed to get userId")
		ErrorJSON(w, "invalid userId", http.StatusBadRequest)
		return
	}

	var input struct {
		OldPassword     string `json:"oldPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	if input.NewPassword != input.ConfirmPassword {
		ErrorJSON(w, "new password do not match", http.StatusBadRequest)
		return
	}

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		ErrorJSON(w, "Failed to parse userId", http.StatusBadRequest)
		return
	}

	var user repository.User
	err = h.Repository.GetUserById(&user, parsedUserId)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "not found user", http.StatusBadRequest)
		return
	}

	if !helper.CheckPasswordHash(user.Password, input.OldPassword) {
		ErrorJSON(w, "invalid password", http.StatusBadRequest)
		return
	}

	hashedPassword, err := helper.HashPassword(input.NewPassword)
	if err != nil {
		InternalServerError(w)
		return
	}

	err = h.Repository.UpdatePassword(parsedUserId, hashedPassword)
	if err != nil {
		log.Error(err)
		InternalServerError(w)
		return
	}

	var out struct {
	}
	ResponseJSON(w, out, http.StatusOK)
}

// UpdateRole godoc
// @Tags Users
// @Summary Update role
// @Description Update role
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} UserJson
// @Router /users/{userId}/role [put]
func (h *APIHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["userId"]
	if !ok {
		log.Error("Failed to get userId")
		ErrorJSON(w, "invalid userId", http.StatusBadRequest)
		return
	}

	var input struct {
		Role string `json:"role"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		ErrorJSON(w, "Failed to parse userId", http.StatusBadRequest)
		return
	}

	var user repository.User
	err = h.Repository.GetUserById(&user, parsedUserId)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "not found user", http.StatusBadRequest)
		return
	}

	err = h.Repository.UpdateRole(parsedUserId, input.Role)
	if err != nil {
		log.Error(err)
		InternalServerError(w)
		return
	}

	var out struct {
		Role string `json:"role"`
	}
	out.Role = input.Role

	ResponseJSON(w, out, http.StatusOK)
}

// Signout godoc
// @Tags Auth
// @Summary Sign out
// @Description Sign out
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} object
// @Router /signout [post]
func (h *APIHandler) Signout(w http.ResponseWriter, r *http.Request) {
	var out struct {
	}

	ResponseJSON(w, out, http.StatusOK)
}

// SigninByToken godoc
// @Tags Auth
// @Summary Sign by token
// @Description Sign by token
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} object
// @Router /token [post]
func (h *APIHandler) SigninByToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusUnauthorized)
		return
	}

	if _, err := helper.VerifyToken(input.Token); err != nil {
		ErrorJSON(w, "failed to verify token", http.StatusUnauthorized)
		return
	}

	var out struct {
		Authorized bool `json:"authorized"`
	}

	out.Authorized = true
	ResponseJSON(w, out, http.StatusOK)
}

// Signup godoc
// @Tags Auth
// @Summary Sign up by administrator
// @Description Sign up by administrator
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} object
// @Router /signup [post]
func (h *APIHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountId string `json:"accountId"`
		Password  string `json:"password"`
		Name      string `json:"name"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	var user repository.User
	err = h.Repository.GetUserByAccountId(&user, input.AccountId)
	if err == nil {
		ErrorJSON(w, "The user existed already", http.StatusBadRequest)
		return
	}

	hashedPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		InternalServerError(w)
		return
	}

	var newUser repository.User
	newUser.AccountId = input.AccountId
	newUser.Password = hashedPassword
	newUser.Name = input.Name
	newUser.Tutorial = false
	newUser.Role = "ADMIN_PROJECT"
	newUser.Initial = true
	err = h.Repository.CreateUser(&newUser)
	if err != nil {
		ErrorJSON(w, "failed to create user", http.StatusBadRequest)
		return
	}

	var out struct {
	}

	ResponseJSON(w, out, http.StatusOK)

}

// GetUsers godoc
// @Tags Users
// @Summary Get users
// @Description Get user list
// @Accept json
// @Produce json
// @Success 200 {object} []UserJson
// @Router /users [get]
func (h *APIHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var out struct {
		Users []UserJson `json:"users"`
	}
	out.Users = make([]UserJson, 0)

	var users []repository.User
	err := h.Repository.GetUsers(&users)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "not found user", http.StatusBadRequest)
		return
	}

	for _, user := range users {
		outUser := UserJson{}

		reflectUser(&outUser, user)
		out.Users = append(out.Users, outUser)
	}

	ResponseJSON(w, out, http.StatusOK)
}

func reflectUser(out *UserJson, user repository.User) {
	out.Id = user.Id.String()
	out.AccountId = user.AccountId
	out.Name = user.Name
	out.Authorized = true
	out.Tutorial = user.Tutorial
	out.Role = user.Role
	out.Initial = user.Initial
	out.CreatedAt = user.CreatedAt
	out.UpdatedAt = user.UpdatedAt
}
