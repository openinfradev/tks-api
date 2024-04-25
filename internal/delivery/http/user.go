package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IUserHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdateUsers(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)

	GetMyProfile(w http.ResponseWriter, r *http.Request)
	UpdateMyProfile(w http.ResponseWriter, r *http.Request)
	UpdateMyPassword(w http.ResponseWriter, r *http.Request)
	RenewPasswordExpiredDate(w http.ResponseWriter, r *http.Request)
	DeleteMyProfile(w http.ResponseWriter, r *http.Request)

	CheckId(w http.ResponseWriter, r *http.Request)
	CheckEmail(w http.ResponseWriter, r *http.Request)
	GetPermissionsByAccountId(w http.ResponseWriter, r *http.Request)

	// Admin
	Admin_Create(w http.ResponseWriter, r *http.Request)
	Admin_Delete(w http.ResponseWriter, r *http.Request)
	Admin_Update(w http.ResponseWriter, r *http.Request)
}

type UserHandler struct {
	usecase           usecase.IUserUsecase
	authUsecase       usecase.IAuthUsecase
	roleUsecase       usecase.IRoleUsecase
	permissionUsecase usecase.IPermissionUsecase
}

func NewUserHandler(h usecase.Usecase) IUserHandler {
	return &UserHandler{
		usecase:           h.User,
		authUsecase:       h.Auth,
		roleUsecase:       h.Role,
		permissionUsecase: h.Permission,
	}
}

// Create godoc
//
//	@Tags			Users
//	@Summary		Create user
//	@Description	Create user
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			body			body		domain.CreateUserRequest	true	"create user request"
//	@Success		200				{object}	domain.CreateUserResponse	"create user response"
//	@Router			/organizations/{organizationId}/users [post]
//	@Security		JWT
func (u UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.CreateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var user model.User
	if err = serializer.Map(r.Context(), input, &user); err != nil {
		log.Error(r.Context(), err)
	}
	user.Organization = model.Organization{
		ID: organizationId,
	}
	for _, role := range input.Roles {
		v, err := u.roleUsecase.GetTksRole(ctx, organizationId, *role.ID)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		user.Roles = append(user.Roles, *v)
	}

	resUser, err := u.usecase.Create(ctx, &user)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusConflict {
			ErrorJSON(w, r, httpErrors.NewConflictError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateUserResponse
	if err = serializer.Map(r.Context(), *resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
	}

	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusCreated, out)

}

// Get godoc
//
//	@Tags			Users
//	@Summary		Get user detail
//	@Description	Get user detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"accountId"
//	@Success		200				{object}	domain.GetUserResponse
//	@Router			/organizations/{organizationId}/users/{accountId} [get]
//	@Security		JWT
func (u UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), userId, organizationId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetUserResponse
	if err = serializer.Map(r.Context(), *user, &out.User); err != nil {
		log.Error(r.Context(), err)
	}

	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusOK, out)
}

// List godoc
//
//	@Tags			Users
//	@Summary		Get user list
//	@Description	Get user list
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"organizationId"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	[]domain.ListUserBody
//	@Router			/organizations/{organizationId}/users [get]
//	@Security		JWT
func (u UserHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	users, err := u.usecase.ListWithPagination(r.Context(), organizationId, pg)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListUserResponse
	out.Users = make([]domain.ListUserBody, len(*users))
	for i, user := range *users {
		if err = serializer.Map(r.Context(), user, &out.Users[i]); err != nil {
			log.Error(r.Context(), err)
		}

		if out.Users[i].Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles); err != nil {
			log.Error(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

func (u UserHandler) convertUserRolesToSimpleRoleResponse(roles []model.Role) []domain.SimpleRoleResponse {
	var simpleRoles []domain.SimpleRoleResponse
	for _, role := range roles {
		simpleRoles = append(simpleRoles, domain.SimpleRoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	return simpleRoles
}

// Delete godoc
//
//	@Tags			Users
//	@Summary		Delete user
//	@Description	Delete user
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"accountId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/users/{accountId} [delete]
//	@Security		JWT
func (u UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "", ""))
		return
	}

	err := u.usecase.DeleteByAccountId(r.Context(), userId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// Update godoc
//
//	@Tags			Users
//	@Summary		Update user
//	@Description	Update user
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			accountId		path		string						true	"accountId"
//	@Param			body			body		domain.UpdateUserRequest	true	"input"
//	@Success		200				{object}	domain.UpdateUserResponse
//	@Router			/organizations/{organizationId}/users/{accountId} [put]
//	@Security		JWT
func (u UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "", ""))
		return
	}

	input := domain.UpdateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var user model.User
	if err = serializer.Map(r.Context(), input, &user); err != nil {
		log.Error(r.Context(), err)
	}
	user.Organization = model.Organization{
		ID: organizationId,
	}
	user.AccountId = accountId
	for _, role := range input.Roles {
		v, err := u.roleUsecase.GetTksRole(ctx, organizationId, *role.ID)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		user.Roles = append(user.Roles, *v)
	}

	resUser, err := u.usecase.UpdateByAccountIdByAdmin(ctx, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateUserResponse
	if err = serializer.Map(r.Context(), *resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}
	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateUsers godoc
//
//	@Tags			Users
//	@Summary		Update multiple users
//	@Description	Update multiple users
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string						true	"organizationId"
//	@Param			body			body	[]domain.UpdateUsersRequest	true	"input"
//	@Success		200
//	@Router			/organizations/{organizationId}/users [put]
//	@Security		JWT
func (u UserHandler) UpdateUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.UpdateUsersRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()

	users := make([]model.User, len(input.Users))
	for i, user := range input.Users {
		if err = serializer.Map(r.Context(), user, &users[i]); err != nil {
			ErrorJSON(w, r, err)
			return
		}
		users[i].Organization = model.Organization{
			ID: organizationId,
		}
		users[i].AccountId = user.AccountId

		for _, role := range user.Roles {
			v, err := u.roleUsecase.GetTksRole(ctx, organizationId, *role.ID)
			if err != nil {
				ErrorJSON(w, r, err)
				return
			}

			users[i].Roles = append(users[i].Roles, *v)
		}

		//ToDo: Implement transaction
		_, err := u.usecase.UpdateByAccountIdByAdmin(r.Context(), &users[i])
		if err != nil {
			if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
				ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
				return
			}

			ErrorJSON(w, r, err)
			return
		}
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// ResetPassword godoc
//
//	@Tags			Users
//	@Summary		Reset user's password as temporary password by admin
//	@Description	Reset user's password as temporary password by admin and send email to user
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string	true	"organizationId"
//	@Param			accountId		path	string	true	"accountId"
//	@Success		200
//	@Router			/organizations/{organizationId}/users/{accountId}/reset-password [put]
//	@Security		JWT
func (u UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	err := u.usecase.ResetPasswordByAccountId(r.Context(), accountId, organizationId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetMyProfile godoc
//
//	@Tags			My-profile
//	@Summary		Get my profile detail
//	@Description	Get my profile detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetMyProfileResponse
//	@Router			/organizations/{organizationId}/my-profile [get]
//	@Security		JWT
func (u UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
		return
	}

	user, err := u.usecase.Get(r.Context(), requestUserInfo.GetUserId())
	if err != nil {
		ErrorJSON(w, r, err)
	}

	var out domain.GetMyProfileResponse
	if err = serializer.Map(r.Context(), *user, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateMyProfile godoc
//
//	@Tags			My-profile
//	@Summary		Update my profile detail
//	@Description	Update my profile detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			body			body		domain.UpdateMyProfileRequest	true	"Required fields: password due to double-check"
//	@Success		200				{object}	domain.UpdateMyProfileResponse
//	@Router			/organizations/{organizationId}/my-profile [put]
//	@Security		JWT
func (u UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
		return
	}

	input := domain.UpdateMyProfileRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	err = u.usecase.ValidateAccount(r.Context(), requestUserInfo.GetUserId(), input.Password, requestUserInfo.GetOrganizationId())
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := u.usecase.Get(r.Context(), requestUserInfo.GetUserId())
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	user.ID = requestUserInfo.GetUserId()
	user.Name = input.Name
	user.Email = input.Email
	user.Department = input.Department

	resUser, err := u.usecase.Update(ctx, user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateMyProfileResponse
	if err = serializer.Map(r.Context(), *resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateMyPassword godoc
//
//	@Tags			My-profile
//	@Summary		Update user password detail
//	@Description	Update user password detail
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string							true	"organizationId"
//	@Param			body			body	domain.UpdatePasswordRequest	true	"update user password request"
//	@Success		200
//	@Router			/organizations/{organizationId}/my-profile/password [put]
//	@Security		JWT
func (u UserHandler) UpdateMyPassword(w http.ResponseWriter, r *http.Request) {
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
		return

	}
	input := domain.UpdatePasswordRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	user, err := u.usecase.Get(r.Context(), requestUserInfo.GetUserId())
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	err = u.usecase.UpdatePasswordByAccountId(r.Context(), user.AccountId, input.OriginPassword, input.NewPassword, requestUserInfo.GetOrganizationId())
	if err != nil {
		if strings.Contains(err.Error(), "invalid origin password") {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "A_INVALID_ORIGIN_PASSWORD", ""))
			return
		}

		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// RenewPasswordExpiredDate godoc
//
//	@Tags			My-profile
//	@Summary		Update user's password expired date to current date
//	@Description	Update user's password expired date to current date
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string	true	"organizationId"
//	@Success		200
//	@Failure		400	{object}	httpErrors.RestError
//	@Router			/organizations/{organizationId}/my-profile/next-password-change [put]
//	@Security		JWT
func (u UserHandler) RenewPasswordExpiredDate(w http.ResponseWriter, r *http.Request) {
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
		return
	}

	err := u.usecase.RenewalPasswordExpiredTime(r.Context(), requestUserInfo.GetUserId())
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteMyProfile godoc
//
//	@Tags			My-profile
//	@Summary		Delete myProfile
//	@Description	Delete myProfile
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string	true	"organizationId"
//	@Success		200
//	@Failure		400
//	@Router			/organizations/{organizationId}/my-profile [delete]
//	@Security		JWT
func (u UserHandler) DeleteMyProfile(w http.ResponseWriter, r *http.Request) {
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request context"), "A_INVALID_TOKEN", ""))
		return
	}
	if err := u.usecase.Delete(r.Context(), requestUserInfo.GetUserId(), requestUserInfo.GetOrganizationId()); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// CheckId godoc
//
//	@Tags			Users
//	@Summary		Get user id existence
//	@Description	return true when accountId exists
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"accountId"
//	@Success		200				{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/users/account-id/{accountId}/existence [get]
//	@Security		JWT
func (u UserHandler) CheckId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	exist := true
	_, err := u.usecase.GetByAccountId(r.Context(), accountId, organizationId)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// CheckEmail godoc
//
//	@Tags			Users
//	@Summary		Get user email existence
//	@Description	return true when email exists
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"email"
//	@Success		200				{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/users/email/{email}/existence [get]
//	@Security		JWT
func (u UserHandler) CheckEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email, ok := vars["email"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	exist := true
	_, err := u.usecase.GetByEmail(r.Context(), email, organizationId)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPermissionsByAccountId godoc
//
//	@Tags			Users
//	@Summary		Get Permissions By Account ID
//	@Description	Get Permissions By Account ID
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			accountId		path		string	true	"Account ID"
//	@Success		200				{object}	domain.GetUsersPermissionsResponse
//	@Router			/organizations/{organizationId}/users/{accountId}/permissions [get]
//	@Security		JWT
func (u UserHandler) GetPermissionsByAccountId(w http.ResponseWriter, r *http.Request) {
	var organizationId, accountId string

	vars := mux.Vars(r)
	if v, ok := vars["accountId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		accountId = v
	}
	if v, ok := vars["organizationId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		organizationId = v
	}

	// ToDo: admin portal에 대해서는 현재 permission 관련 로직을 적용하지 않아 임시로 빈 값을 리턴하도록 함
	if organizationId == "master" {
		ResponseJSON(w, r, http.StatusOK, domain.GetUsersPermissionsResponse{})
		return
	}

	user, err := u.usecase.GetByAccountId(r.Context(), accountId, organizationId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	var permissionSets []*model.PermissionSet
	for _, role := range user.Roles {
		permissionSet, err := u.permissionUsecase.GetPermissionSetByRoleId(r.Context(), role.ID)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		permissionSets = append(permissionSets, permissionSet)
	}

	mergedPermissionSet := u.permissionUsecase.MergePermissionWithOrOperator(r.Context(), permissionSets...)

	var out domain.GetUsersPermissionsResponse
	out.Permissions = make([]*domain.MergePermissionResponse, 0)
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Dashboard))
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Stack))
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Policy))
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.ProjectManagement))
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Notification))
	out.Permissions = append(out.Permissions, convertModelToMergedPermissionSetResponse(r.Context(), mergedPermissionSet.Configuration))

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertModelToMergedPermissionSetResponse(ctx context.Context, permission *model.Permission) *domain.MergePermissionResponse {
	var permissionResponse domain.MergePermissionResponse

	permissionResponse.Key = permission.Key
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToMergedPermissionSetResponse(ctx, child))
	}

	return &permissionResponse
}

// Admin_Create godoc
//
//	@Tags			Users
//	@Summary		Create user by admin in Admin Portal
//	@Description	Create user by admin in Admin Portal
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			body			body		domain.Admin_CreateUserRequest	true	"create user request"
//	@Success		200				{object}	domain.Admin_CreateUserResponse	"create user response"
//	@Router			/admin/organizations/{organizationId}/users [post]
//	@Security		JWT
func (u UserHandler) Admin_Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.Admin_CreateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	// TKS 관리자가 아닌 경우 Password 확인
	if organizationId != "master" {
		// check admin password
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
			return
		}
		err = u.usecase.ValidateAccount(r.Context(), requestUserInfo.GetUserId(), input.AdminPassword, requestUserInfo.GetOrganizationId())
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}
	}

	user := model.User{
		Name:        input.Name,
		AccountId:   input.AccountId,
		Email:       input.Email,
		Department:  input.Department,
		Description: input.Description,
	}

	for _, role := range input.Roles {
		v, err := u.roleUsecase.GetTksRole(r.Context(), organizationId, *role.ID)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}

		user.Roles = append(user.Roles, *v)
	}

	user.Organization = model.Organization{
		ID: organizationId,
	}

	user.Password = u.usecase.GenerateRandomPassword(r.Context())

	resUser, err := u.usecase.Create(r.Context(), &user)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusConflict {
			ErrorJSON(w, r, httpErrors.NewConflictError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	err = u.usecase.SendEmailForTemporaryPassword(r.Context(), user.AccountId, organizationId, user.Password)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.Admin_CreateUserResponse

	out.ID = resUser.ID.String()
	ResponseJSON(w, r, http.StatusCreated, out)

}

// Admin_Delete godoc
//
//	@Tags			Users
//	@Summary		Delete user by admin in Admin Portal
//	@Description	Delete user by admin in Admin Portal
//	@Accept			json
//	@Produce		json
//	@Param			body			body	domain.DeleteUserRequest	true	"input"
//	@Param			organizationId	path	string						true	"organizationId"
//	@Param			accountId		path	string						true	"accountId"
//	@Success		200
//	@Router			/admin/organizations/{organizationId}/users/{accountId} [delete]
//	@Security		JWT
func (u UserHandler) Admin_Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "", ""))
		return
	}

	input := domain.DeleteUserRequest{}

	// TKS 관리자가 아닌 경우 Password 확인
	if organizationId != "master" {
		err := UnmarshalRequestInput(r, &input)
		if err != nil {
			log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

			ErrorJSON(w, r, err)
			return
		}

		// check admin password
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
			return
		}
		err = u.usecase.ValidateAccount(r.Context(), requestUserInfo.GetUserId(), input.AdminPassword, requestUserInfo.GetOrganizationId())
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}
	}

	err := u.usecase.DeleteByAccountId(r.Context(), userId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// Admin_Update godoc
//
//	@Tags			Users
//	@Summary		Update user by admin in Admin Portal
//	@Description	Update user by admin in Admin Portal
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			accountId		path		string							true	"accountId"
//	@Param			body			body		domain.Admin_UpdateUserRequest	true	"input"
//	@Success		200				{object}	domain.Admin_UpdateUserResponse
//	@Router			/admin/organizations/{organizationId}/users/{accountId} [put]
//	@Security		JWT
func (u UserHandler) Admin_Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId, ok := vars["accountId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("accountId not found in path"), "C_INVALID_ACCOUNT_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "", ""))
		return
	}

	input := domain.Admin_UpdateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	// TKS 관리자가 아닌 경우 Password 확인
	if organizationId != "master" {
		// check admin password
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found in request"), "A_INVALID_TOKEN", ""))
			return
		}
		err = u.usecase.ValidateAccount(r.Context(), requestUserInfo.GetUserId(), input.AdminPassword, requestUserInfo.GetOrganizationId())
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}
	}

	ctx := r.Context()
	user := model.User{
		AccountId:   accountId,
		Name:        input.Name,
		Email:       input.Email,
		Department:  input.Department,
		Description: input.Description,
	}
	user.Organization = model.Organization{
		ID: organizationId,
	}
	for _, role := range input.Roles {
		user.Roles = append(user.Roles, model.Role{
			ID: *role.ID,
		})
	}

	resUser, err := u.usecase.UpdateByAccountIdByAdmin(ctx, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.Admin_UpdateUserResponse
	if err = serializer.Map(r.Context(), *resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	out.User.Roles = u.convertUserRolesToSimpleRoleResponse(user.Roles)

	ResponseJSON(w, r, http.StatusOK, out)
}
