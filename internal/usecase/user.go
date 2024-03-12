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
	CreateAdmin(ctx context.Context, organizationId string, accountId string, accountName string, email string) (*model.User, error)
	DeleteAdmin(ctx context.Context, organizationId string) error
	DeleteAll(ctx context.Context, organizationId string) error
	Create(ctx context.Context, user *model.User) (*model.User, error)
	List(ctx context.Context, organizationId string) (*[]model.User, error)
	ListWithPagination(ctx context.Context, organizationId string, pg *pagination.Pagination) (*[]model.User, error)
	Get(ctx context.Context, userId uuid.UUID) (*model.User, error)
	Update(ctx context.Context, userId uuid.UUID, user *model.User) (*model.User, error)
	ResetPassword(ctx context.Context, userId uuid.UUID) error
	ResetPasswordByAccountId(ctx context.Context, accountId string, organizationId string) error
	GenerateRandomPassword(ctx context.Context, ) string
	Delete(ctx context.Context, userId uuid.UUID, organizationId string) error
	GetByAccountId(ctx context.Context, accountId string, organizationId string) (*model.User, error)
	GetByEmail(ctx context.Context, email string, organizationId string) (*model.User, error)
	SendEmailForTemporaryPassword(ctx context.Context, accountId string, organizationId string, password string) error

	UpdateByAccountId(ctx context.Context, accountId string, user *model.User) (*model.User, error)
	UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string, organizationId string) error
	RenewalPasswordExpiredTime(ctx context.Context, userId uuid.UUID) error
	RenewalPasswordExpiredTimeByAccountId(ctx context.Context, accountId string, organizationId string) error
	DeleteByAccountId(ctx context.Context, accountId string, organizationId string) error
	ValidateAccount(ctx context.Context, userId uuid.UUID, password string, organizationId string) error
	ValidateAccountByAccountId(ctx context.Context, accountId string, password string, organizationId string) error

	UpdateByAccountIdByAdmin(ctx context.Context, accountId string, user *model.User) (*model.User, error)
}

type UserUsecase struct {
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
		log.ErrorfWithContext(ctx, "failed to update password expired time: %v", err)
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
	userInKeycloak, err := u.kc.GetUser(user.Organization.ID, user.AccountId)
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
	if err = u.kc.UpdateUser(user.Organization.ID, userInKeycloak); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	if err = u.userRepository.UpdatePasswordAt(ctx, userId, user.Organization.ID, true); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	message, err := mail.MakeTemporaryPasswordMessage(user.Email, user.Organization.ID, user.AccountId, randomPassword)
	if err != nil {
		log.Errorf("mail.MakeVerityIdentityMessage error. %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(); err != nil {
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
	_, err = u.kc.Login(user.AccountId, password, organizationId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid password"), "A_INVALID_PASSWORD", "")
	}
	return nil
}

func (u *UserUsecase) ValidateAccountByAccountId(ctx context.Context, accountId string, password string, organizationId string) error {
	_, err := u.kc.Login(organizationId, accountId, password)
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

	err = u.userRepository.DeleteWithUuid(ctx, userUuid)
	if err != nil {
		return errors.Wrap(err, "delete user failed")
	}

	return nil
}

func (u *UserUsecase) CreateAdmin(ctx context.Context, organizationId string, accountId string, accountName string, email string) (*model.User, error) {
	// Generate Admin user object
	randomPassword := helper.GenerateRandomString(passwordLength)
	user := model.User{
		AccountId: accountId,
		Password:  randomPassword,
		Email:     email,
		Role: model.Role{
			Name: "admin",
		},
		Organization: model.Organization{
			ID: organizationId,
		},
		Name: accountName,
	}

	// Create Admin user in keycloak & DB
	resUser, err := u.Create(context.Background(), &user)
	if err != nil {
		return nil, err
	}

	// Send mail of temporary password
	organizationInfo, err := u.organizationRepository.Get(ctx, organizationId)
	if err != nil {
		return nil, err
	}
	message, err := mail.MakeGeneratingOrganizationMessage(organizationId, organizationInfo.Name, user.Email, user.AccountId, randomPassword)
	if err != nil {
		return nil, httpErrors.NewInternalServerError(err, "", "")
	}
	mailer := mail.New(message)
	if err := mailer.SendMail(); err != nil {
		return nil, httpErrors.NewInternalServerError(err, "", "")
	}

	return resUser, nil
}

func (u *UserUsecase) SendEmailForTemporaryPassword(ctx context.Context, accountId string, organizationId string, password string) error {
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		return err
	}

	message, err := mail.MakeTemporaryPasswordMessage(user.Email, organizationId, accountId, password)
	if err != nil {
		return err
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(); err != nil {
		return err
	}

	return nil

}

func (u *UserUsecase) UpdatePasswordByAccountId(ctx context.Context, accountId string, originPassword string, newPassword string,
	organizationId string) error {
	if originPassword == newPassword {
		return httpErrors.NewBadRequestError(fmt.Errorf("new password is same with origin password"), "A_SAME_OLD_PASSWORD", "")
	}
	if _, err := u.kc.Login(accountId, originPassword, organizationId); err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid origin password"), "A_INVALID_PASSWORD", "")
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

func (u *UserUsecase) Update(ctx context.Context, userId uuid.UUID, user *model.User) (*model.User, error) {
	storedUser, err := u.Get(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.AccountId = storedUser.AccountId

	return u.UpdateByAccountId(ctx, storedUser.AccountId, user)
}

func (u *UserUsecase) UpdateByAccountId(ctx context.Context, accountId string, user *model.User) (*model.User, error) {
	var out model.User

	originUser, err := u.kc.GetUser(user.Organization.ID, accountId)
	if err != nil {
		return nil, err
	}

	if (originUser.Email == nil || *originUser.Email != user.Email) || (originUser.FirstName == nil || *originUser.FirstName != user.Name) {
		originUser.Email = gocloak.StringP(user.Email)
		originUser.FirstName = gocloak.StringP(user.Name)
		err = u.kc.UpdateUser(user.Organization.ID, originUser)
		if err != nil {
			return nil, err
		}
	}

	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(user.Organization.ID),
		u.userRepository.AccountIdFilter(accountId))
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

	roleUuid := (*users)[0].Role.ID
	if err != nil {
		return nil, err
	}

	out, err = u.userRepository.UpdateWithUuid(ctx, (*users)[0].ID, user.AccountId, user.Name, roleUuid, user.Email, user.Department, user.Description)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return &out, nil
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
	err = u.kc.DeleteUser(organizationId, user.AccountId)
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
	err = u.kc.DeleteUser(organizationId, accountId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) Create(ctx context.Context, user *model.User) (*model.User, error) {
	// Create user in keycloak
	groups := []string{fmt.Sprintf("%s@%s", user.Role.Name, user.Organization.ID)}
	userUuidStr, err := u.kc.CreateUser(user.Organization.ID, &gocloak.User{
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

	// Get user role
	var roleUuid string
	roles, err := u.roleRepository.ListTksRoles(ctx, user.Organization.ID, nil)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, httpErrors.NewInternalServerError(fmt.Errorf("role not found"), "", "")
	}
	for _, role := range roles {
		if role.Name == user.Role.Name {
			roleUuid = role.ID
		}
	}
	if roleUuid == "" {
		return nil, httpErrors.NewInternalServerError(fmt.Errorf("role not found"), "", "")
	}

	// Generate user uuid
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return nil, err
	}

	// Create user in DB
	resUser, err := u.userRepository.CreateWithUuid(ctx, userUuid, user.AccountId, user.Name, user.Email,
		user.Department, user.Description, user.Organization.ID, roleUuid)
	if err != nil {
		return nil, err
	}

	return &resUser, nil
}

func (u *UserUsecase) UpdateByAccountIdByAdmin(ctx context.Context, accountId string, newUser *model.User) (*model.User, error) {
	deepCopyUser := *newUser
	user, err := u.UpdateByAccountId(ctx, accountId, &deepCopyUser)
	if err != nil {
		return nil, err
	}

	if newUser.Role.Name != user.Role.Name {
		originGroupName := fmt.Sprintf("%s@%s", user.Role.Name, newUser.Organization.ID)
		newGroupName := fmt.Sprintf("%s@%s", newUser.Role.Name, newUser.Organization.ID)
		if err := u.kc.LeaveGroup(newUser.Organization.ID, user.ID.String(), originGroupName); err != nil {
			log.ErrorfWithContext(ctx, "leave group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "", "")
		}
		if err := u.kc.JoinGroup(newUser.Organization.ID, user.ID.String(), newGroupName); err != nil {
			log.ErrorfWithContext(ctx, "join group in keycloak failed: %v", err)
			return nil, httpErrors.NewInternalServerError(err, "", "")
		}
	}

	*user, err = u.userRepository.UpdateWithUuid(ctx, user.ID, user.AccountId, user.Name, newUser.Role.ID, user.Email,
		user.Department, user.Description)
	if err != nil {
		return nil, errors.Wrap(err, "updating user in repository failed")
	}

	return user, nil
}

func NewUserUsecase(r repository.Repository, kc keycloak.IKeycloak) IUserUsecase {
	return &UserUsecase{
		userRepository:         r.User,
		roleRepository:         r.Role,
		kc:                     kc,
		organizationRepository: r.Organization,
	}
}
