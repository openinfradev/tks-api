package usecase

import (
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IRoleUsecase interface {
	CreateTksRole(role *model.Role) (string, error)
	ListRoles(pg *pagination.Pagination) ([]*model.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	GetTksRole(id string) (*model.Role, error)
	DeleteTksRole(id string) error
	UpdateTksRole(role *model.Role) error
}

type RoleUsecase struct {
	repo repository.IRoleRepository
}

func NewRoleUsecase(repo repository.Repository) *RoleUsecase {
	return &RoleUsecase{
		repo: repo.Role,
	}
}

func (r RoleUsecase) CreateTksRole(role *model.Role) (string, error) {
	return r.repo.Create(role)
}

func (r RoleUsecase) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*model.Role, error) {
	roles, err := r.repo.ListTksRoles(organizationId, pg)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r RoleUsecase) ListRoles(pg *pagination.Pagination) ([]*model.Role, error) {
	return r.repo.List(nil)
}

func (r RoleUsecase) GetTksRole(id string) (*model.Role, error) {
	role, err := r.repo.GetTksRole(id)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) DeleteTksRole(id string) error {
	return r.repo.Delete(id)
}

func (r RoleUsecase) UpdateTksRole(role *model.Role) error {
	err := r.repo.Update(role)
	if err != nil {
		return err
	}

	return nil
}
