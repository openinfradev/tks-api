package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal/aws/ses"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type IUserUsecase interface {
	CreateAdmin(organizationId string) (*domain.User, error)
	DeleteAdmin(organizationId string) error
	DeleteAll(ctx context.Context, organizationId string) error
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	List(ctx context.Context, organizationId string) (*[]domain.User, error)
	Get(userId uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, userId uuid.UUID, user *domain.User) (*domain.User, error)
	ResetPassword(userId uuid.UUID) error
	ResetPasswordByAccountId(accountId string, organizationId string) error
	Delete(userId uuid.UUID, organizationId string) error
	GetByAccountId(ctx context.Context, accountId string, organizationId string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string, organizationId string) (*domain.User, error)

	UpdateByAccountId(ctx context.Context, accountId string, user *domain.User) (*domain.User, error)
	UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string, organizationId string) error
	RenewalPasswordExpiredTime(ctx context.Context, userId uuid.UUID) error
	RenewalPasswordExpiredTimeByAccountId(ctx context.Context, accountId string, organizationId string) error
	DeleteByAccountId(ctx context.Context, accountId string, organizationId string) error
	ValidateAccount(userId uuid.UUID, password string, organizationId string) error
	ValidateAccountByAccountId(accountId string, password string, organizationId string) error

	UpdateByAccountIdByAdmin(ctx context.Context, accountId string, user *domain.User) (*domain.User, error)
}

type UserUsecase struct {
	repo repository.IUserRepository
	kc   keycloak.IKeycloak
}

func (u *UserUsecase) RenewalPasswordExpiredTime(ctx context.Context, userId uuid.UUID) error {
	user, err := u.repo.GetByUuid(userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status != http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
		return httpErrors.NewInternalServerError(err, "")
	}

	err = u.repo.UpdatePassword(userId, user.Organization.ID, user.Password, false)
	if err != nil {
		log.Errorf("failed to update password expired time: %v", err)
		return httpErrors.NewInternalServerError(err, "")
	}

	return nil
}

func (u *UserUsecase) RenewalPasswordExpiredTimeByAccountId(ctx context.Context, accountId string, organizationId string) error {
	user, err := u.repo.Get(accountId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status != http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
		return httpErrors.NewInternalServerError(err, "")
	}
	userId, err := uuid.Parse(user.ID)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}
	return u.RenewalPasswordExpiredTime(ctx, userId)
}

func (u *UserUsecase) ResetPassword(userId uuid.UUID) error {
	user, err := u.repo.GetByUuid(userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
	}
	userInKeycloak, err := u.kc.GetUser(user.Organization.ID, user.AccountId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
		return httpErrors.NewInternalServerError(err, "")
	}

	randomPassword := helper.GenerateRandomString(passwordLength)
	userInKeycloak.Credentials = &[]gocloak.CredentialRepresentation{
		{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(randomPassword),
			Temporary: gocloak.BoolP(false),
		},
	}
	if err = u.kc.UpdateUser(user.Organization.ID, userInKeycloak); err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}

	if user.Password, err = helper.HashPassword(randomPassword); err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}
	if err = u.repo.UpdatePassword(userId, user.Organization.ID, user.Password, true); err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}

	if err = ses.SendEmailForTemporaryPassword(ses.Client, user.Email, randomPassword); err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}

	return nil
}

func (u *UserUsecase) ResetPasswordByAccountId(accountId string, organizationId string) error {
	user, err := u.repo.Get(accountId, organizationId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
		return httpErrors.NewInternalServerError(err, "")
	}
	userId, err := uuid.Parse(user.ID)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "")
	}
	return u.ResetPassword(userId)
}

func (u *UserUsecase) ValidateAccount(userId uuid.UUID, password string, organizationId string) error {
	user, err := u.repo.GetByUuid(userId)
	if err != nil {
		return err
	}
	_, err = u.kc.Login(user.AccountId, password, organizationId)
	return err
}

func (u *UserUsecase) ValidateAccountByAccountId(accountId string, password string, organizationId string) error {
	_, err := u.kc.Login(organizationId, accountId, password)
	return err
}

func (u *UserUsecase) DeleteAll(ctx context.Context, organizationId string) error {
	// TODO: implement me as transaction
	// TODO: clean users in keycloak

	err := u.repo.Flush(organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) DeleteAdmin(organizationId string) error {
	user, err := u.kc.GetUser(organizationId, "admin")
	if err != nil {
		return errors.Wrap(err, "get user failed")
	}

	err = u.kc.DeleteUser(organizationId, "admin")
	if err != nil {
		return errors.Wrap(err, "delete user failed")
	}

	userUuid, err := uuid.Parse(*user.ID)
	if err != nil {
		return errors.Wrap(err, "parse user id failed")
	}

	err = u.repo.DeleteWithUuid(userUuid)
	if err != nil {
		return errors.Wrap(err, "delete user failed")
	}

	return nil
}

func (u *UserUsecase) CreateAdmin(orgainzationId string) (*domain.User, error) {
	user := domain.User{
		AccountId: "admin",
		Password:  "admin",
		Role: domain.Role{
			Name: "admin",
		},
		Organization: domain.Organization{
			ID: orgainzationId,
		},
		Name: "admin",
	}

	// Create user in keycloak
	groups := []string{fmt.Sprintf("%s@%s", user.Role.Name, orgainzationId)}
	err := u.kc.CreateUser(orgainzationId, &gocloak.User{
		Username: gocloak.StringP(user.AccountId),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(user.Password),
				Temporary: gocloak.BoolP(false),
			},
		},
		Groups: &groups,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating user in keycloak failed")
	}
	keycloakUser, err := u.kc.GetUser(user.Organization.ID, user.AccountId)
	if err != nil {
		return nil, errors.Wrap(err, "getting user from keycloak failed")
	}

	userUuid, err := uuid.Parse(*keycloakUser.ID)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := helper.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	roles, err := u.repo.FetchRoles()
	if err != nil {
		return nil, err
	}
	for _, role := range *roles {
		if role.Name == user.Role.Name {
			user.Role.ID = role.ID
		}
	}
	roleUuid, err := uuid.Parse(user.Role.ID)
	if err != nil {
		return nil, err
	}
	resUser, err := u.repo.CreateWithUuid(userUuid, user.AccountId, user.Name, hashedPassword, user.Email,
		user.Department, user.Description, user.Organization.ID, roleUuid)
	if err != nil {
		return nil, err
	}

	return &resUser, nil
}

func (u *UserUsecase) UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string,
	organizationId string) error {
	if originPassword == newPassword {
		return httpErrors.NewBadRequestError(fmt.Errorf("new password is same with origin password"), "")
	}
	if _, err := u.kc.Login(accountId, originPassword, organizationId); err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid origin password"), "")
	}
	originUser, err := u.kc.GetUser(organizationId, accountId)
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

	err = u.kc.UpdateUser(organizationId, originUser)
	if err != nil {
		return errors.Wrap(err, "updating user in keycloak failed")
	}

	// update password in DB

	user, err := u.repo.Get(accountId, organizationId)
	if err != nil {
		return errors.Wrap(err, "getting user from repository failed")
	}
	userUuid, err := uuid.Parse(user.ID)
	if err != nil {
		return errors.Wrap(err, "parsing uuid failed")
	}
	hashedPassword, err := helper.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err, "hashing password failed")
	}

	err = u.repo.UpdatePassword(userUuid, organizationId, hashedPassword, false)
	if err != nil {
		return errors.Wrap(err, "updating user in repository failed")
	}

	return nil
}

func (u *UserUsecase) List(ctx context.Context, organizationId string) (*[]domain.User, error) {
	users, err := u.repo.List(u.repo.OrganizationFilter(organizationId))
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *UserUsecase) Get(userId uuid.UUID) (*domain.User, error) {
	user, err := u.repo.GetByUuid(userId)
	if err != nil {
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			return nil, httpErrors.NewBadRequestError(fmt.Errorf("user not found"), "")
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserUsecase) GetByAccountId(ctx context.Context, accountId string, organizationId string) (*domain.User, error) {
	users, err := u.repo.List(u.repo.OrganizationFilter(organizationId),
		u.repo.AccountIdFilter(accountId))
	if err != nil {
		return nil, err
	}

	return &(*users)[0], nil
}

func (u *UserUsecase) GetByEmail(ctx context.Context, email string, organizationId string) (*domain.User, error) {
	users, err := u.repo.List(u.repo.OrganizationFilter(organizationId),
		u.repo.EmailFilter(email))
	if err != nil {
		return nil, err
	}

	return &(*users)[0], nil
}

func (u *UserUsecase) Update(ctx context.Context, userId uuid.UUID, user *domain.User) (*domain.User, error) {
	storedUser, err := u.Get(userId)
	if err != nil {
		return nil, err
	}
	user.AccountId = storedUser.AccountId

	return u.UpdateByAccountId(ctx, storedUser.AccountId, user)
}

func (u *UserUsecase) UpdateByAccountId(ctx context.Context, accountId string, user *domain.User) (*domain.User, error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("user in the context is  empty")
	}

	_, err := u.kc.Login(user.AccountId, user.Password, userInfo.GetOrganizationId())
	if err != nil {
		return nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid password"), "")
	}

	originUser, err := u.kc.GetUser(userInfo.GetOrganizationId(), accountId)
	if err != nil {
		return nil, err
	}
	if originUser.Email == nil || *originUser.Email != user.Email {
		originUser.Email = gocloak.StringP(user.Email)
		err = u.kc.UpdateUser(userInfo.GetOrganizationId(), originUser)
		if err != nil {
			return nil, err
		}
	}

	users, err := u.repo.List(u.repo.OrganizationFilter(userInfo.GetOrganizationId()),
		u.repo.AccountIdFilter(accountId))
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			return nil, httpErrors.NewNotFoundError(httpErrors.NotFound, "")
		}
		return nil, errors.Wrap(err, "getting users from repository failed")
	}
	if len(*users) == 0 {
		return nil, fmt.Errorf("user not found")
	} else if len(*users) > 1 {
		return nil, fmt.Errorf("multiple users found")
	}

	userUuid, err := uuid.Parse((*users)[0].ID)
	if err != nil {
		return nil, err
	}

	originPassword := (*users)[0].Password

	roleUuid, err := uuid.Parse((*users)[0].Role.ID)
	if err != nil {
		return nil, err
	}

	*user, err = u.repo.UpdateWithUuid(userUuid, user.AccountId, user.Name, originPassword, roleUuid, user.Email,
		user.Department, user.Description)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return user, nil
}

func (u *UserUsecase) Delete(userId uuid.UUID, organizationId string) error {
	user, err := u.repo.GetByUuid(userId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("not found user"), "")
	}

	err = u.repo.DeleteWithUuid(userId)
	if err != nil {
		return err
	}

	// Delete user in keycloak
	err = u.kc.DeleteUser(organizationId, user.AccountId)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserUsecase) DeleteByAccountId(ctx context.Context, accountId string, organizationId string) error {
	user, err := u.repo.Get(accountId, organizationId)
	if err != nil {
		return err
	}

	userUuid, err := uuid.Parse(user.ID)
	if err != nil {
		return err
	}
	err = u.repo.DeleteWithUuid(userUuid)
	if err != nil {
		return err
	}

	// Delete user in keycloak
	err = u.kc.DeleteUser(organizationId, accountId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Create user in keycloak
	groups := []string{fmt.Sprintf("%s@%s", user.Role.Name, user.Organization.ID)}
	err := u.kc.CreateUser(user.Organization.ID, &gocloak.User{
		Username: gocloak.StringP(user.AccountId),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(user.Password),
				Temporary: gocloak.BoolP(false),
			},
		},
		Email:  gocloak.StringP(user.Email),
		Groups: &groups,
	})
	if err != nil {
		if _, err := u.kc.GetUser(user.Organization.ID, user.AccountId); err == nil {
			return nil, httpErrors.NewConflictError(errors.New("user already exists"), "")
		}

		return nil, errors.Wrap(err, "creating user in keycloak failed")
	}
	keycloakUser, err := u.kc.GetUser(user.Organization.ID, user.AccountId)
	if err != nil {
		return nil, errors.Wrap(err, "getting user from keycloak failed")
	}

	userUuid, err := uuid.Parse(*keycloakUser.ID)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := helper.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	roles, err := u.repo.FetchRoles()
	if err != nil {
		return nil, err
	}
	for _, role := range *roles {
		if role.Name == user.Role.Name {
			user.Role.ID = role.ID
		}
	}
	roleUuid, err := uuid.Parse(user.Role.ID)
	if err != nil {
		return nil, err
	}

	resUser, err := u.repo.CreateWithUuid(userUuid, user.AccountId, user.Name, hashedPassword, user.Email,
		user.Department, user.Description, user.Organization.ID, roleUuid)
	if err != nil {
		return nil, err
	}

	return &resUser, nil
}

func (u *UserUsecase) UpdateByAccountIdByAdmin(ctx context.Context, accountId string, user *domain.User) (*domain.User, error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("user in the context is  empty")
	}

	originUser, err := u.kc.GetUser(userInfo.GetOrganizationId(), accountId)
	if err != nil {
		return nil, err
	}
	if originUser.Email == nil || *originUser.Email != user.Email {
		originUser.Email = gocloak.StringP(user.Email)
		err = u.kc.UpdateUser(userInfo.GetOrganizationId(), originUser)
		if err != nil {
			return nil, err
		}
	}

	users, err := u.repo.List(u.repo.OrganizationFilter(userInfo.GetOrganizationId()),
		u.repo.AccountIdFilter(accountId))
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			return nil, httpErrors.NewNotFoundError(httpErrors.NotFound, "")
		}
		return nil, errors.Wrap(err, "getting users from repository failed")
	}
	if len(*users) == 0 {
		return nil, fmt.Errorf("user not found")
	} else if len(*users) > 1 {
		return nil, fmt.Errorf("multiple users found")
	}
	if user.Role.Name != (*users)[0].Role.Name {
		originGroupName := fmt.Sprintf("%s@%s", (*users)[0].Role.Name, userInfo.GetOrganizationId())
		newGroupName := fmt.Sprintf("%s@%s", user.Role.Name, userInfo.GetOrganizationId())
		if err := u.kc.LeaveGroup(userInfo.GetOrganizationId(), (*users)[0].ID, originGroupName); err != nil {
			log.Errorf("leave group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "")
		}
		if err := u.kc.JoinGroup(userInfo.GetOrganizationId(), (*users)[0].ID, newGroupName); err != nil {
			log.Errorf("join group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "")
		}
	}

	userUuid, err := uuid.Parse((*users)[0].ID)
	if err != nil {
		return nil, err
	}

	originPassword := (*users)[0].Password

	roles, err := u.repo.FetchRoles()
	if err != nil {
		return nil, err
	}
	for _, role := range *roles {
		if role.Name == user.Role.Name {
			user.Role.ID = role.ID
		}
	}
	roleUuid, err := uuid.Parse(user.Role.ID)
	if err != nil {
		return nil, err
	}

	*user, err = u.repo.UpdateWithUuid(userUuid, user.AccountId, user.Name, originPassword, roleUuid, user.Email,
		user.Department, user.Description)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return user, nil
}

func NewUserUsecase(r repository.Repository, kc keycloak.IKeycloak) IUserUsecase {
	return &UserUsecase{
		repo: r.User,
		kc:   kc,
	}
}
