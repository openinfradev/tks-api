package usecase

import (
	"fmt"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IAuthUsecase interface {
	Signin(accountId string, password string) (domain.User, error)
	Register(accountId string, password string, name string) (domain.User, error)
}

type AuthUsecase struct {
	repo repository.IAuthRepository
}

func NewAuthUsecase(r repository.IAuthRepository) IAuthUsecase {
	return &AuthUsecase{
		repo: r,
	}
}

func (r *AuthUsecase) Signin(accountId string, password string) (domain.User, error) {
	user, err := r.repo.GetUserByAccountId(accountId)
	if err != nil {
		return domain.User{}, err
	}

	if !helper.CheckPasswordHash(user.Password, password) {
		log.Debug(user.Password)
		log.Debug(password)
		return domain.User{}, fmt.Errorf("Invalid password")
	}

	user.Token, err = helper.CreateJWT(accountId, user.ID)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to create token")
	}

	return user, nil
}

func (r *AuthUsecase) Register(accountId string, password string, name string) (domain.User, error) {
	_, err := r.repo.GetUserByAccountId(accountId)
	if err == nil {
		return domain.User{}, fmt.Errorf("Already existed user. %s", accountId)
	}

	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	resUser, err := r.repo.Create(accountId, hashedPassword, name)
	if err != nil {
		return domain.User{}, err
	}

	// [TODO] 임시로 tks-admin 으로 세팅한다.
	err = r.repo.AssignRole(accountId, "tks-admin")
	if err != nil {
		return domain.User{}, err
	}

	return resUser, nil
}
