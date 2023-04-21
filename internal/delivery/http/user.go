package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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

	UpdateByAdmin(w http.ResponseWriter, r *http.Request)
}

type UserHandler struct {
	usecase usecase.IUserUsecase
}

func NewUserHandler(h usecase.IUserUsecase) IUserHandler {
	return &UserHandler{
		usecase: h,
	}
}

// Create godoc
// @Tags Users
// @Summary Create user
// @Description Create user
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param body body domain.CreateUserRequest true "create user request"
// @Success 200 {object} domain.CreateUserResponse "create user response"
// @Router /organizations/{organizationId}/users [post]
// @Security     JWT
func (u UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	input := domain.CreateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	//userInfo, ok := request.UserFrom(r.Context())
	//if !ok {
	//	ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("user info not found in token")))
	//	return
	//}

	ctx := r.Context()
	var user domain.User
	if err = domain.Map(input, &user); err != nil {
		log.Error(err)
	}
	user.Organization = domain.Organization{
		ID: organizationId,
	}

	resUser, err := u.usecase.Create(ctx, &user)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusConflict {
			ErrorJSON(w, httpErrors.NewConflictError(err))
			return
		}

		ErrorJSON(w, err)
		return
	}

	var out domain.CreateUserResponse
	if err = domain.Map(*resUser, &out.User); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusCreated, out)

}

// Get godoc
// @Tags Users
// @Summary Get user detail
// @Description Get user detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Success 200 {object} domain.GetUserResponse
// @Router /organizations/{organizationId}/users/{accountId} [get]
// @Security     JWT
func (u UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId, organizationId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}

		ErrorJSON(w, err)
		return
	}

	var out domain.GetUserResponse
	if err = domain.Map(*user, &out.User); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// List godoc
// @Tags Users
// @Summary Get user list
// @Description Get user list
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} []domain.ListUserBody
// @Router /organizations/{organizationId}/users [get]
// @Security     JWT
func (u UserHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	users, err := u.usecase.List(r.Context(), organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ResponseJSON(w, http.StatusNoContent, domain.ListUserResponse{})
			return
		}

		log.Errorf("error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, err)
		return
	}

	var out domain.ListUserResponse
	out.Users = make([]domain.ListUserBody, len(*users))
	for i, user := range *users {
		if err = domain.Map(user, &out.Users[i]); err != nil {
			log.Error(err)
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// Delete godoc
// @Tags Users
// @Summary Delete user
// @Description Delete user
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Success 200 {object} domain.User
// @Router /organizations/{organizationId}/users/{accountId} [delete]
// @Security     JWT
func (u UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	err := u.usecase.DeleteByAccountId(r.Context(), userId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// Update UpdateUser godoc
// @Tags Users
// @Summary Update user detail
// @Description Update user detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Param body body domain.UpdateUserRequest true "update user request"
// @Success 200 {object} domain.UpdateUserResponse
// @Router /organizations/{organizationId}/users/{accountId} [put]
// @Security     JWT
func (u UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	input := domain.UpdateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	ctx := r.Context()
	var user domain.User
	if err = domain.Map(input, &user); err != nil {
		log.Error(err)
	}
	user.Organization = domain.Organization{
		ID: organizationId,
	}
	user.AccountId = accountId
	resUser, err := u.usecase.UpdateByAccountId(ctx, accountId, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
		}

		ErrorJSON(w, err)
		return
	}

	var out domain.UpdateUserResponse
	if err = domain.Map(*resUser, &out.User); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// UpdatePassword godoc
// @Tags Users
// @Summary Update user password detail
// @Description Update user password detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Param body body domain.UpdatePasswordRequest true "update user password request"
// @Success 200 {object} domain.UpdatePasswordResponse
// @Router /organizations/{organizationId}/users/{accountId}/password [put]
// @Security     JWT
func (u UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	input := domain.UpdatePasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	err = u.usecase.UpdatePasswordByAccountId(r.Context(), accountId, input.OriginPassword, input.NewPassword, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
			return
		}
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// CheckId godoc
// @Tags Users
// @Summary Get user id existence
// @Description return true when accountId exists
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Success 200
// @Router /organizations/{organizationId}/users/{accountId}/existence [get]
// @Security     JWT
func (u UserHandler) CheckId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	exist := true
	_, err := u.usecase.GetByAccountId(r.Context(), accountId, organizationId)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, err)
			return
		}
	}

	var out domain.CheckExistedIdResponse
	out.Existed = exist

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateByAdmin UpdateUser godoc
// @Tags Admin
// @Summary As admin, Update user detail
// @Description Update user detail
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param accountId path string true "accountId"
// @Param body body domain.UpdateUserByAdminRequest true "update user request"
// @Success 200 {object} domain.UpdateUserByAdminResponse
// @Router /admin/organizations/{organizationId}/users/{accountId} [put]
// @Security     JWT
func (u UserHandler) UpdateByAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path")))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path")))
		return
	}

	input := domain.UpdateUserByAdminRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	ctx := r.Context()
	var user domain.User
	if err = domain.Map(input, &user); err != nil {
		log.Error(err)
	}
	user.Organization = domain.Organization{
		ID: organizationId,
	}
	user.AccountId = accountId
	user.Role.Name = input.Role

	resUser, err := u.usecase.UpdateByAccountIdByAdmin(ctx, accountId, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, httpErrors.NewBadRequestError(err))
		}

		ErrorJSON(w, err)
		return
	}

	var out domain.UpdateUserResponse
	if err = domain.Map(*resUser, &out.User); err != nil {
		log.Error(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}
