package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/mail"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type IUserUsecase interface {
	CreateAdmin(ctx context.Context, user *model.User) (*model.User, error)
	DeleteAdmin(ctx context.Context, organizationId string) error
	DeleteAll(ctx context.Context, organizationId string) error
	Create(ctx context.Context, user *model.User) (*model.User, error)
	List(ctx context.Context, organizationId string) (*[]model.User, error)
	ListWithPagination(ctx context.Context, organizationId string, pg *pagination.Pagination) (*[]model.User, error)
	Get(ctx context.Context, userId uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	ResetPassword(ctx context.Context, userId uuid.UUID) error
	ResetPasswordByAccountId(ctx context.Context, accountId string, organizationId string) error
	GenerateRandomPassword(ctx context.Context) string
	Delete(ctx context.Context, userId uuid.UUID, organizationId string) error
	GetByAccountId(ctx context.Context, accountId string, organizationId string) (*model.User, error)
	GetByEmail(ctx context.Context, email string, organizationId string) (*model.User, error)
	SendEmailForTemporaryPassword(ctx context.Context, accountId string, organizationId string, password string) error

	UpdateByAccountId(ctx context.Context, user *model.User) (*model.User, error)
	UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string, organizationId string) error
	RenewalPasswordExpiredTime(ctx context.Context, userId uuid.UUID) error
	RenewalPasswordExpiredTimeByAccountId(ctx context.Context, accountId string, organizationId string) error
	DeleteByAccountId(ctx context.Context, accountId string, organizationId string) error
	ValidateAccount(ctx context.Context, userId uuid.UUID, password string, organizationId string) error
	ValidateAccountByAccountId(ctx context.Context, accountId string, password string, organizationId string) error

	UpdateByAccountIdByAdmin(ctx context.Context, user *model.User) (*model.User, error)
}

type UserUsecase struct {
	authRepository         repository.IAuthRepository
	userRepository         repository.IUserRepository
	roleRepository         repository.IRoleRepository
	organizationRepository repository.IOrganizationRepository
	kc                     keycloak.IKeycloak
}

func (u *UserUsecase) RenewalPasswordExpiredTime(ctx context.Context, userId uuid.UUID) error {
	user, err := u.userRepository.GetByUuid(ctx, userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status != http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
		return httpErrors.NewInternalServerError(err, "", "")
	}

	err = u.userRepository.UpdatePasswordAt(ctx, userId, user.Organization.ID, false)
	if err != nil {
		log.Errorf(ctx, "failed to update password expired time: %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (u *UserUsecase) RenewalPasswordExpiredTimeByAccountId(ctx context.Context, accountId string, organizationId string) error {
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status != http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
		return httpErrors.NewInternalServerError(err, "", "")
	}
	return u.RenewalPasswordExpiredTime(ctx, user.ID)
}

func (u *UserUsecase) ResetPassword(ctx context.Context, userId uuid.UUID) error {
	user, err := u.userRepository.GetByUuid(ctx, userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
	}
	userInKeycloak, err := u.kc.GetUser(ctx, user.Organization.ID, user.AccountId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
		return httpErrors.NewInternalServerError(err, "", "")
	}

	randomPassword := helper.GenerateRandomString(passwordLength)
	userInKeycloak.Credentials = &[]gocloak.CredentialRepresentation{
		{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(randomPassword),
			Temporary: gocloak.BoolP(false),
		},
	}
	if err = u.kc.UpdateUser(ctx, user.Organization.ID, userInKeycloak); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	if err = u.userRepository.UpdatePasswordAt(ctx, userId, user.Organization.ID, true); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	message, err := mail.MakeTemporaryPasswordMessage(ctx, user.Email, user.Organization.ID, user.AccountId, randomPassword)
	if err != nil {
		log.Errorf(ctx, "mail.MakeVerityIdentityMessage error. %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(ctx); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (u *UserUsecase) ResetPasswordByAccountId(ctx context.Context, accountId string, organizationId string) error {
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
		return httpErrors.NewInternalServerError(err, "", "")
	}
	return u.ResetPassword(ctx, user.ID)
}

func (u *UserUsecase) GenerateRandomPassword(ctx context.Context) string {
	return helper.GenerateRandomString(passwordLength)
}

func (u *UserUsecase) ValidateAccount(ctx context.Context, userId uuid.UUID, password string, organizationId string) error {
	user, err := u.userRepository.GetByUuid(ctx, userId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
	}
	_, err = u.kc.Login(ctx, user.AccountId, password, organizationId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid password"), "A_INVALID_PASSWORD", "")
	}
	return nil
}

func (u *UserUsecase) ValidateAccountByAccountId(ctx context.Context, accountId string, password string, organizationId string) error {
	_, err := u.kc.Login(ctx, organizationId, accountId, password)
	return err
}

func (u *UserUsecase) DeleteAll(ctx context.Context, organizationId string) error {
	// TODO: implement me as transaction
	// TODO: clean users in keycloak

	err := u.userRepository.Flush(ctx, organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) DeleteAdmin(ctx context.Context, organizationId string) error {
	user, err := u.kc.GetUser(ctx, organizationId, "admin")
	if err != nil {
		return errors.Wrap(err, "get user failed")
	}

	err = u.kc.DeleteUser(ctx, organizationId, "admin")
	if err != nil {
		return errors.Wrap(err, "delete user failed")
	}

	userUuid, err := uuid.Parse(*user.ID)
	if err != nil {
		return errors.Wrap(err, "parse user id failed")
	}

	err = u.userRepository.DeleteWithUuid(ctx, userUuid)
	if err != nil {
		return errors.Wrap(err, "delete user failed")
	}

	return nil
}

func (u *UserUsecase) CreateAdmin(ctx context.Context, user *model.User) (*model.User, error) {
	// Generate Admin user object
	randomPassword := helper.GenerateRandomString(passwordLength)
	user.Password = randomPassword

	// Create Admin user in keycloak & DB
	resUser, err := u.Create(context.Background(), user)
	if err != nil {
		return nil, err
	}

	// Send mail of temporary password
	organizationInfo, err := u.organizationRepository.Get(ctx, resUser.Organization.ID)
	if err != nil {
		return nil, err
	}
	message, err := mail.MakeGeneratingOrganizationMessage(ctx, resUser.Organization.ID, organizationInfo.Name, user.Email, user.AccountId, randomPassword)
	if err != nil {
		return nil, httpErrors.NewInternalServerError(err, "", "")
	}
	mailer := mail.New(message)
	if err := mailer.SendMail(ctx); err != nil {
		return nil, httpErrors.NewInternalServerError(err, "", "")
	}

	return resUser, nil
}

func (u *UserUsecase) SendEmailForTemporaryPassword(ctx context.Context, accountId string, organizationId string, password string) error {
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		return err
	}

	message, err := mail.MakeTemporaryPasswordMessage(ctx, user.Email, organizationId, accountId, password)
	if err != nil {
		return err
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(ctx); err != nil {
		return err
	}

	return nil

}

func (u *UserUsecase) UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string,
	organizationId string) error {
	if originPassword == newPassword {
		return httpErrors.NewBadRequestError(fmt.Errorf("new password is same with origin password"), "A_SAME_OLD_PASSWORD", "")
	}
	if _, err := u.kc.Login(ctx, accountId, originPassword, organizationId); err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid origin password"), "A_INVALID_PASSWORD", "")
	}
	originUser, err := u.kc.GetUser(ctx, organizationId, accountId)
	if err != nil {
		return err
	}
	originUser.Credentials = &[]gocloak.CredentialRepresentation{
		{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(newPassword),
			Temporary: gocloak.BoolP(false),
		},
	}

	err = u.kc.UpdateUser(ctx, organizationId, originUser)
	if err != nil {
		return errors.Wrap(err, "updating user in keycloak failed")
	}

	// update password UpdateAt in DB
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		return errors.Wrap(err, "getting user from repository failed")
	}

	err = u.userRepository.UpdatePasswordAt(ctx, user.ID, organizationId, false)
	if err != nil {
		return errors.Wrap(err, "updating user in repository failed")
	}

	return nil
}

func (u *UserUsecase) List(ctx context.Context, organizationId string) (users *[]model.User, err error) {
	users, err = u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId))
	if err != nil {
		return nil, err
	}

	return
}

func (u *UserUsecase) ListWithPagination(ctx context.Context, organizationId string, pg *pagination.Pagination) (users *[]model.User, err error) {
	users, err = u.userRepository.ListWithPagination(ctx, pg, organizationId)
	if err != nil {
		return nil, err
	}

	return
}

func (u *UserUsecase) Get(ctx context.Context, userId uuid.UUID) (*model.User, error) {
	user, err := u.userRepository.GetByUuid(ctx, userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return nil, httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "U_NO_USER", "")
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserUsecase) GetByAccountId(ctx context.Context, accountId string, organizationId string) (*model.User, error) {
	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.AccountIdFilter(accountId))
	if err != nil {
		return nil, err
	}

	return &(*users)[0], nil
}

func (u *UserUsecase) GetByEmail(ctx context.Context, email string, organizationId string) (*model.User, error) {
	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.EmailFilter(email))
	if err != nil {
		return nil, err
	}

	return &(*users)[0], nil
}

func (u *UserUsecase) Update(ctx context.Context, user *model.User) (*model.User, error) {
	storedUser, err := u.Get(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.AccountId = storedUser.AccountId

	return u.UpdateByAccountId(ctx, user)
}

func (u *UserUsecase) UpdateByAccountId(ctx context.Context, user *model.User) (*model.User, error) {
	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(user.Organization.ID),
		u.userRepository.AccountIdFilter(user.AccountId))
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			return nil, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
		}
		return nil, errors.Wrap(err, "getting users from repository failed")
	}
	if len(*users) == 0 {
		return nil, fmt.Errorf("user not found")
	} else if len(*users) > 1 {
		return nil, fmt.Errorf("multiple users found")
	}

	if ((*users)[0].Email != user.Email) || ((*users)[0].Name != user.Name) {
		err = u.kc.UpdateUser(ctx, user.Organization.ID, &gocloak.User{
			Email:     gocloak.StringP(user.Email),
			FirstName: gocloak.StringP(user.Name),
		})
		if err != nil {
			return nil, err
		}
	}

	resp, err := u.userRepository.Update(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return resp, nil
}

func (u *UserUsecase) Delete(ctx context.Context, userId uuid.UUID, organizationId string) error {
	user, err := u.userRepository.GetByUuid(ctx, userId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("not found user"), "", "")
	}

	err = u.userRepository.DeleteWithUuid(ctx, userId)
	if err != nil {
		return err
	}

	// Delete user in keycloak
	err = u.kc.DeleteUser(ctx, organizationId, user.AccountId)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserUsecase) DeleteByAccountId(ctx context.Context, accountId string, organizationId string) error {
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		return err
	}

	err = u.userRepository.DeleteWithUuid(ctx, user.ID)
	if err != nil {
		return err
	}

	// Delete user in keycloak
	err = u.kc.DeleteUser(ctx, organizationId, accountId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) Create(ctx context.Context, user *model.User) (*model.User, error) {
	// Create user in keycloak
	var groups []string
	for _, role := range user.Roles {
		groups = append(groups, fmt.Sprintf("%s@%s", role.Name, user.Organization.ID))
	}

	userUuidStr, err := u.kc.CreateUser(ctx, user.Organization.ID, &gocloak.User{
		Username: gocloak.StringP(user.AccountId),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(user.Password),
				Temporary: gocloak.BoolP(false),
			},
		},
		Email:     gocloak.StringP(user.Email),
		Groups:    &groups,
		FirstName: gocloak.StringP(user.Name),
	})
	if err != nil {
		return nil, err
	}

	if user.ID, err = uuid.Parse(userUuidStr); err != nil {
		return nil, err
	}

	// Create user in DB
	resUser, err := u.userRepository.Create(ctx, user)
	//resUser, err := u.userRepository.Create(ctx, userUuid, user.AccountId, user.Name, user.Email,
	//	user.Department, user.Description, user.Organization.ID, roleUuid)
	if err != nil {
		return nil, err
	}

	return resUser, nil
}

func (u *UserUsecase) UpdateByAccountIdByAdmin(ctx context.Context, newUser *model.User) (*model.User, error) {
	if newUser.AccountId == "" {
		return nil, httpErrors.NewBadRequestError(fmt.Errorf("accountId is required"), "C_INVALID_ACCOUNT_ID", "")
	}

	deepCopyUser := *newUser
	user, err := u.UpdateByAccountId(ctx, &deepCopyUser)
	if err != nil {
		return nil, err
	}

	var unassigningRoleIds, assigningRoleIds map[string]struct{}
	for _, role := range newUser.Roles {
		assigningRoleIds[role.ID] = struct{}{}
	}
	for _, role := range user.Roles {
		if _, ok := assigningRoleIds[role.ID]; !ok {
			unassigningRoleIds[role.ID] = struct{}{}
		} else {
			delete(assigningRoleIds, role.ID)
		}
	}

	for roleId := range unassigningRoleIds {
		groupName := fmt.Sprintf("%s@%s", roleId, user.Organization.ID)
		if err := u.kc.LeaveGroup(ctx, user.Organization.ID, user.ID.String(), groupName); err != nil {
			log.Errorf(ctx, "leave group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "", "")
		}
	}

	for roleId := range assigningRoleIds {
		groupName := fmt.Sprintf("%s@%s", roleId, user.Organization.ID)
		if err := u.kc.JoinGroup(ctx, user.Organization.ID, user.ID.String(), groupName); err != nil {
			log.Errorf(ctx, "join group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "", "")
		}
	}

	err = u.authRepository.UpdateExpiredTimeOnToken(ctx, user.Organization.ID, user.ID.String())

	user.Roles = newUser.Roles
	resp, err := u.userRepository.Update(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return resp, nil
}

func NewUserUsecase(r repository.Repository, kc keycloak.IKeycloak) IUserUsecase {
	return &UserUsecase{
		authRepository:         r.Auth,
		userRepository:         r.User,
		roleRepository:         r.Role,
		kc:                     kc,
		organizationRepository: r.Organization,
	}
}
