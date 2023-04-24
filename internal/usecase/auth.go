package usecase

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuthUsecase interface {
	Login(accountId string, password string, organizationId string) (domain.User, error)
	Logout(sessionId string, orgainzationId string) error
	FetchRoles() (out []domain.Role, err error)
}

type AuthUsecase struct {
	kc   keycloak.IKeycloak
	repo repository.IUserRepository
}

func NewAuthUsecase(r repository.Repository, kc keycloak.IKeycloak) IAuthUsecase {
	return &AuthUsecase{
		repo: r.User,
		kc:   kc,
	}
}

func (r *AuthUsecase) Login(accountId string, password string, organizationId string) (domain.User, error) {
	// Authentication with DB
	user, err := r.repo.Get(accountId, organizationId)
	if err != nil {
		return domain.User{}, httpErrors.NewUnauthorizedError(err)
	}
	if !helper.CheckPasswordHash(user.Password, password) {
		return domain.User{}, httpErrors.NewUnauthorizedError(fmt.Errorf(""))
	}
	var accountToken *domain.User
	// Authentication with Keycloak
	if organizationId == "master" && accountId == "admin" {
		accountToken, err = r.kc.LoginAdmin(accountId, password)
	} else {
		accountToken, err = r.kc.Login(accountId, password, organizationId)
	}
	if err != nil {
		//TODO: implement not found handling
		return domain.User{}, httpErrors.NewUnauthorizedError(err)
	}

	// Insert token
	user.Token = accountToken.Token

	return user, nil
}

func (r *AuthUsecase) Logout(sessionId string, organizationId string) error {
	err := r.kc.Logout(sessionId, organizationId)
	if err != nil {
		return err
	}
	return nil
}

func (u *AuthUsecase) FetchRoles() (out []domain.Role, err error) {
	roles, err := u.repo.FetchRoles()
	if err != nil {
		return nil, err
	}
	return *roles, nil
}
