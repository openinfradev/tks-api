package usecase

import (
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/aws/ses"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuthUsecase interface {
	Login(accountId string, password string, organizationId string) (domain.User, error)
	Logout(accessToken string, organizationId string) error
	FindId(code string, email string, userName string, organizationId string) (string, error)
	FindPassword(code string, accountId string, email string, userName string, organizationId string) error
	VerifyIdentity(accountId string, email string, userName string, organizationId string) error
	FetchRoles() (out []domain.Role, err error)
}

const (
	passwordLength = 8
)

type AuthUsecase struct {
	kc             keycloak.IKeycloak
	userRepository repository.IUserRepository
	authRepository repository.IAuthRepository
}

func NewAuthUsecase(r repository.Repository, kc keycloak.IKeycloak) IAuthUsecase {
	return &AuthUsecase{
		userRepository: r.User,
		authRepository: r.Auth,
		kc:             kc,
	}
}

func (u *AuthUsecase) Login(accountId string, password string, organizationId string) (domain.User, error) {
	// Authentication with DB
	user, err := u.userRepository.Get(accountId, organizationId)
	if err != nil {
		return domain.User{}, httpErrors.NewUnauthorizedError(err)
	}
	if !helper.CheckPasswordHash(user.Password, password) {
		return domain.User{}, httpErrors.NewUnauthorizedError(fmt.Errorf(""))
	}
	var accountToken *domain.User
	// Authentication with Keycloak
	if organizationId == "master" && accountId == "admin" {
		accountToken, err = u.kc.LoginAdmin(accountId, password)
	} else {
		accountToken, err = u.kc.Login(accountId, password, organizationId)
	}
	if err != nil {
		//TODO: implement not found handling
		return domain.User{}, httpErrors.NewUnauthorizedError(err)
	}

	// Insert token
	user.Token = accountToken.Token

	user.PasswordExpired = helper.IsDurationExpired(user.PasswordUpdatedAt, internal.PasswordExpiredDuration)

	return user, nil
}

func (u *AuthUsecase) Logout(accessToken string, organizationName string) error {
	// [TODO] refresh token 을 추가하고, session timeout 을 줄이는 방향으로 고려할 것
	err := u.kc.Logout(accessToken, organizationName)
	if err != nil {
		return err
	}
	return nil
}
func (u *AuthUsecase) FindId(code string, email string, userName string, organizationId string) (string, error) {
	users, err := u.userRepository.List(u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.NameFilter(userName), u.userRepository.EmailFilter(email))
	if err != nil && users == nil {
		return "", httpErrors.NewBadRequestError(err)
	}
	if err != nil {
		return "", httpErrors.NewInternalServerError(err)
	}
	userUuid, err := uuid.Parse((*users)[0].ID)
	if err != nil {
		return "", httpErrors.NewInternalServerError(err)
	}
	emailCode, err := u.authRepository.GetEmailCode(userUuid)
	if err != nil {
		return "", httpErrors.NewBadRequestError(err)
	}
	if !u.isValidEmailCode(emailCode) {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("invalid code"))
	}
	if emailCode.Code != code {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("invalid code"))
	}
	if err := u.authRepository.DeleteEmailCode(userUuid); err != nil {
		return "", httpErrors.NewInternalServerError(err)
	}

	return (*users)[0].AccountId, nil
}

func (u *AuthUsecase) FindPassword(code string, accountId string, email string, userName string, organizationId string) error {
	users, err := u.userRepository.List(u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.AccountIdFilter(accountId), u.userRepository.NameFilter(userName),
		u.userRepository.EmailFilter(email))
	if err != nil && users == nil {
		return httpErrors.NewBadRequestError(err)
	}
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	user := (*users)[0]
	userUuid, err := uuid.Parse(user.ID)
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	emailCode, err := u.authRepository.GetEmailCode(userUuid)
	if err != nil {
		return httpErrors.NewBadRequestError(err)
	}
	if !u.isValidEmailCode(emailCode) {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid code"))
	}
	if emailCode.Code != code {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid code"))
	}
	randomPassword := helper.GenerateRandomString(passwordLength)

	originUser, err := u.kc.GetUser(organizationId, accountId)
	if err != nil {
		return err
	}
	originUser.Credentials = &[]gocloak.CredentialRepresentation{
		{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(randomPassword),
			Temporary: gocloak.BoolP(false),
		},
	}
	if err = u.kc.UpdateUser(organizationId, originUser); err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	if user.Password, err = helper.HashPassword(randomPassword); err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	if err = u.userRepository.UpdatePassword(userUuid, organizationId, user.Password, true); err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	if err = ses.SendEmailForTemporaryPassword(ses.Client, email, randomPassword); err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	if err = u.authRepository.DeleteEmailCode(userUuid); err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	return nil
}

func (u *AuthUsecase) VerifyIdentity(accountId string, email string, userName string, organizationId string) error {
	var users *[]domain.User
	var err error

	if accountId == "" {
		users, err = u.userRepository.List(u.userRepository.OrganizationFilter(organizationId),
			u.userRepository.NameFilter(userName), u.userRepository.EmailFilter(email))
	} else {
		users, err = u.userRepository.List(u.userRepository.OrganizationFilter(organizationId),
			u.userRepository.AccountIdFilter(accountId), u.userRepository.NameFilter(userName),
			u.userRepository.EmailFilter(email))
	}
	if err != nil && users == nil {
		return httpErrors.NewBadRequestError(err)
	}
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	code, err := helper.GenerateEmailCode()
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	userUuid, err := uuid.Parse((*users)[0].ID)
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	_, err = u.authRepository.GetEmailCode(userUuid)
	if err != nil {
		if err := u.authRepository.CreateEmailCode(userUuid, code); err != nil {
			return httpErrors.NewInternalServerError(err)
		}
	} else {
		if err := u.authRepository.UpdateEmailCode(userUuid, code); err != nil {
			return httpErrors.NewInternalServerError(err)
		}
	}
	if err := ses.SendEmailForVerityIdentity(ses.Client, email, code); err != nil {
		return httpErrors.NewInternalServerError(err)
	}

	return nil
}

func (u *AuthUsecase) FetchRoles() (out []domain.Role, err error) {
	roles, err := u.userRepository.FetchRoles()
	if err != nil {
		return nil, err
	}
	return *roles, nil
}

func (u *AuthUsecase) isValidEmailCode(code repository.CacheEmailCode) bool {
	return !helper.IsDurationExpired(code.UpdatedAt, internal.EmailCodeExpireTime)
}
