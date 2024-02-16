package usecase

import (
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IRoleUsecase interface {
	CreateTksRole(role *domain.Role) error
	ListRoles(pg *pagination.Pagination) ([]*domain.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.Role, error)
	GetTksRole(id string) (*domain.Role, error)
	DeleteTksRole(id string) error
	UpdateTksRole(role *domain.Role) error
}

type RoleUsecase struct {
	repo repository.IRoleRepository
}

func NewRoleUsecase(repo repository.Repository) *RoleUsecase {
	return &RoleUsecase{
		repo: repo.Role,
	}
}

func (r RoleUsecase) CreateTksRole(role *domain.Role) error {
	return r.repo.Create(role)
}

func (r RoleUsecase) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.Role, error) {
	roles, err := r.repo.ListTksRoles(organizationId, pg)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r RoleUsecase) ListRoles(pg *pagination.Pagination) ([]*domain.Role, error) {
	return r.repo.List(nil)
}

func (r RoleUsecase) GetTksRole(id string) (*domain.Role, error) {
	role, err := r.repo.GetTksRole(id)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) DeleteTksRole(id string) error {
	return r.repo.Delete(id)
}

func (r RoleUsecase) UpdateTksRole(role *domain.Role) error {
	err := r.repo.Update(role)
	if err != nil {
		return err
	}

	return nil
}
