package usecase

import (
	"context"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IRoleUsecase interface {
	CreateTksRole(ctx context.Context, role *model.Role) (string, error)
	ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	GetTksRole(ctx context.Context, id string) (*model.Role, error)
	DeleteTksRole(ctx context.Context, organizationId string, id string) error
	UpdateTksRole(ctx context.Context, role *model.Role) error
	IsRoleNameExisted(ctx context.Context, organizationId string, roleName string) (bool, error)
}

type RoleUsecase struct {
	repo repository.IRoleRepository
	kc   keycloak.IKeycloak
}

func NewRoleUsecase(repo repository.Repository, kc keycloak.IKeycloak) *RoleUsecase {
	return &RoleUsecase{
		repo: repo.Role,
		kc:   kc,
	}
}

func (r RoleUsecase) CreateTksRole(ctx context.Context, role *model.Role) (string, error) {
	roleId, err := r.kc.CreateGroup(ctx, role.OrganizationID, role.Name)
	if err != nil {
		return "", err
	}
	role.ID = roleId
	return r.repo.Create(ctx, role)
}

func (r RoleUsecase) ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error) {
	roles, err := r.repo.ListTksRoles(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r RoleUsecase) GetTksRole(ctx context.Context, id string) (*model.Role, error) {
	role, err := r.repo.GetTksRole(ctx, id)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) DeleteTksRole(ctx context.Context, organizationId string, id string) error {
	err := r.kc.DeleteGroup(ctx, organizationId, id)
	if err != nil {
		return err
	}
	return r.repo.Delete(ctx, id)
}

func (r RoleUsecase) UpdateTksRole(ctx context.Context, role *model.Role) error {
	err := r.repo.Update(ctx, role)
	if err != nil {
		return err
	}

	return nil
}

func (r RoleUsecase) IsRoleNameExisted(ctx context.Context, organizationId string, roleName string) (bool, error) {
	role, err := r.repo.GetTksRoleByRoleName(ctx, organizationId, roleName)
	if err != nil {
		return false, err
	}

	if role != nil {
		return true, nil
	}

	return false, nil
}
