package http

import (
	"fmt"
	"net/http"
	"strings"

	admin_domain "github.com/openinfradev/tks-api/pkg/domain/admin"

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
	ResetPassword(w http.ResponseWriter, r *http.Request)

	GetMyProfile(w http.ResponseWriter, r *http.Request)
	UpdateMyProfile(w http.ResponseWriter, r *http.Request)
	UpdateMyPassword(w http.ResponseWriter, r *http.Request)
	RenewPasswordExpiredDate(w http.ResponseWriter, r *http.Request)
	DeleteMyProfile(w http.ResponseWriter, r *http.Request)

	CheckId(w http.ResponseWriter, r *http.Request)
	CheckEmail(w http.ResponseWriter, r *http.Request)

	// Admin
	Admin_Create(w http.ResponseWriter, r *http.Request)
	Admin_List(w http.ResponseWriter, r *http.Request)
	Admin_Get(w http.ResponseWriter, r *http.Request)
	Admin_Delete(w http.ResponseWriter, r *http.Request)
	Admin_Update(w http.ResponseWriter, r *http.Request)
}

type UserHandler struct {
	usecase     usecase.IUserUsecase
	authUsecase usecase.IAuthUsecase
	roleUsecase usecase.IRoleUsecase
}

func NewUserHandler(h usecase.Usecase) IUserHandler {
	return &UserHandler{
		usecase:     h.User,
		authUsecase: h.Auth,
		roleUsecase: h.Role,
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var user model.User
	if err = serializer.Map(input, &user); err != nil {
		log.Error(r.Context(), err)
	}
	user.Organization = model.Organization{
		ID: organizationId,
	}

	roles, err := u.roleUsecase.ListTksRoles(r.Context(), organizationId, nil)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	for _, role := range roles {
		if role.Name == input.Role {
			user.Role = *role
			break
		}
	}

	resUser, err := u.usecase.Create(ctx, &user)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusConflict {
			ErrorJSON(w, r, httpErrors.NewConflictError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateUserResponse
	if err = serializer.Map(*resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
	}

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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetUserResponse
	if err = serializer.Map(*user, &out.User); err != nil {
		log.Error(r.Context(), err)
	}

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
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListUserResponse
	out.Users = make([]domain.ListUserBody, len(*users))
	for i, user := range *users {
		if err = serializer.Map(user, &out.Users[i]); err != nil {
			log.Error(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var user model.User
	if err = serializer.Map(input, &user); err != nil {
		ErrorJSON(w, r, err)
		return
	}
	user.Organization = model.Organization{
		ID: organizationId,
	}
	user.AccountId = accountId

	roles, err := u.roleUsecase.ListTksRoles(r.Context(), organizationId, nil)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	for _, role := range roles {
		if role.Name == input.Role {
			user.Role = *role
			break
		}
	}

	resUser, err := u.usecase.UpdateByAccountIdByAdmin(ctx, accountId, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateUserResponse
	if err = serializer.Map(*resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
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
	if err = serializer.Map(*user, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

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
	organizationId, ok := vars["organizationId"]
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	err = u.usecase.ValidateAccount(r.Context(), requestUserInfo.GetUserId(), input.Password, requestUserInfo.GetOrganizationId())
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ctx := r.Context()
	var user model.User
	if err = serializer.Map(input, &user); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	user.OrganizationId = organizationId
	resUser, err := u.usecase.Update(ctx, requestUserInfo.GetUserId(), &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}

	var out domain.UpdateMyProfileResponse
	if err = serializer.Map(*resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

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

// Admin_Create godoc
//	@Tags			Admin
//	@Summary		Create user by admin
//	@Description	Create user by admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			body			body		admin_domain.CreateUserRequest	true	"create user request"
//	@Success		200				{object}	admin_domain.CreateUserResponse	"create user response"
//	@Router			/admin/organizations/{organizationId}/users [post]
//	@Security		JWT

func (u UserHandler) Admin_Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := admin_domain.CreateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

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

	roles, err := u.roleUsecase.ListTksRoles(r.Context(), organizationId, nil)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	for _, role := range roles {
		if role.Name == input.Role {
			user.Role = *role
			break
		}
	}

	user.Organization = model.Organization{
		ID: organizationId,
	}

	user.Password = u.usecase.GenerateRandomPassword(r.Context())

	resUser, err := u.usecase.Create(r.Context(), &user)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusConflict {
			ErrorJSON(w, r, httpErrors.NewConflictError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	err = u.usecase.SendEmailForTemporaryPassword(r.Context(), user.AccountId, organizationId, user.Password)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out admin_domain.CreateUserResponse

	out.ID = resUser.ID.String()
	ResponseJSON(w, r, http.StatusCreated, out)

}

// Admin_List godoc
//	@Tags			Admin
//	@Summary		Get user list by admin
//	@Description	Get user list by admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			limit			query		string							false	"pageSize"
//	@Param			page			query		string							false	"pageNumber"
//	@Param			soertColumn		query		string							false	"sortColumn"
//	@Param			sortOrder		query		string							false	"sortOrder"
//	@Param			filters			query		[]string						false	"filters"
//	@Success		200				{object}	admin_domain.ListUserResponse	"user list response"
//	@Router			/admin/organizations/{organizationId}/users [get]
//	@Security		JWT

func (u UserHandler) Admin_List(w http.ResponseWriter, r *http.Request) {
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var out admin_domain.ListUserResponse
	out.Users = make([]domain.ListUserBody, len(*users))
	for i, user := range *users {
		if err = serializer.Map(user, &out.Users[i]); err != nil {
			log.Error(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// Admin_Get godoc
//
//	@Tags			Admin
//	@Summary		Get user detail by admin
//	@Description	Get user detail by admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"accountId"
//	@Success		200				{object}	admin_domain.GetUserResponse
//	@Router			/admin/organizations/{organizationId}/users/{accountId} [get]
func (u UserHandler) Admin_Get(w http.ResponseWriter, r *http.Request) {
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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out admin_domain.GetUserResponse
	if err = serializer.Map(*user, &out.User); err != nil {
		log.Error(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// Admin_Delete godoc
//	@Tags			Admin
//	@Summary		Delete user by admin
//	@Description	Delete user by admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			accountId		path		string	true	"accountId"
//	@Success		200				{object}	admin_domain.User
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

	input := admin_domain.DeleteUserRequest{}

	// TKS 관리자가 아닌 경우 Password 확인
	if organizationId != "master" {
		err := UnmarshalRequestInput(r, &input)
		if err != nil {
			log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

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
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// Admin_Update godoc
//
//	@Tags			Admin
//	@Summary		Update user by admin
//	@Description	Update user by admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"organizationId"
//	@Param			accountId		path		string							true	"accountId"
//	@Param			body			body		admin_domain.UpdateUserRequest	true	"input"
//	@Success		200				{object}	admin_domain.UpdateUserResponse
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

	input := admin_domain.UpdateUserRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

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

	roles, err := u.roleUsecase.ListTksRoles(r.Context(), organizationId, nil)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}
	for _, role := range roles {
		if role.Name == input.Role {
			user.Role = *role
			break
		}
	}

	resUser, err := u.usecase.UpdateByAccountIdByAdmin(ctx, accountId, &user)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out admin_domain.UpdateUserResponse
	if err = serializer.Map(*resUser, &out.User); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
