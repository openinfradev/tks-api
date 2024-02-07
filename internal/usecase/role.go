package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IRoleUsecase interface {
	CreateTksRole(role *domain.TksRole) error
	CreateProjectRole(role *domain.ProjectRole) error
	ListRoles(pg *pagination.Pagination) ([]*domain.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.TksRole, error)
	ListProjectRoles(projectId string, pg *pagination.Pagination) ([]*domain.ProjectRole, error)
	GetTksRole(id string) (*domain.TksRole, error)
	GetProjectRole(id string) (*domain.ProjectRole, error)
	DeleteTksRole(id string) error
	DeleteProjectRole(id string) error
	UpdateTksRole(role *domain.TksRole) error
	UpdateProjectRole(role *domain.ProjectRole) error
}

type RoleUsecase struct {
	repo repository.IRoleRepository
}

func NewRoleUsecase(repo repository.Repository) *RoleUsecase {
	return &RoleUsecase{
		repo: repo.Role,
	}
}

func (r RoleUsecase) CreateTksRole(role *domain.TksRole) error {
	return r.repo.Create(role)
}

func (r RoleUsecase) CreateProjectRole(role *domain.ProjectRole) error {
	return r.repo.Create(role)
}

func (r RoleUsecase) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.TksRole, error) {
	roles, err := r.repo.ListTksRoles(organizationId, pg)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r RoleUsecase) ListProjectRoles(projectId string, pg *pagination.Pagination) ([]*domain.ProjectRole, error) {
	roles, err := r.repo.ListProjectRoles(projectId, pg)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r RoleUsecase) ListRoles(pg *pagination.Pagination) ([]*domain.Role, error) {
	return r.repo.List(nil)
}

func (r RoleUsecase) GetTksRole(id string) (*domain.TksRole, error) {
	roldId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	role, err := r.repo.GetTksRole(roldId)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) GetProjectRole(id string) (*domain.ProjectRole, error) {
	roleId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	role, err := r.repo.GetProjectRole(roleId)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) DeleteTksRole(id string) error {
	roleId, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return r.repo.DeleteCascade(roleId)
}

func (r RoleUsecase) DeleteProjectRole(id string) error {
	roleId, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return r.repo.DeleteCascade(roleId)
}

func (r RoleUsecase) UpdateTksRole(role *domain.TksRole) error {
	err := r.repo.Update(role)
	if err != nil {
		return err
	}

	return nil
}

func (r RoleUsecase) UpdateProjectRole(role *domain.ProjectRole) error {
	err := r.repo.Update(role)
	if err != nil {
		return err
	}

	return nil
}
