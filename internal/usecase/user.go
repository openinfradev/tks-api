package usecase

import (
	"github.com/gofrs/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IUserUsecase interface {
	Get(userId uuid.UUID) (user domain.User, err error)
	GetByAccountId(accountId string, password string) (user domain.User, err error)
}

type UserUsecase struct {
	repo repository.IUserRepository
}

func NewUserUsecase(r repository.IUserRepository) IUserUsecase {
	return &UserUsecase{
		repo: r,
	}
}

func (r *UserUsecase) Get(userId uuid.UUID) (res domain.User, err error) {
	return domain.User{}, nil
}

func (r *UserUsecase) GetByAccountId(accountId string, password string) (user domain.User, err error) {
	return domain.User{}, nil
}
