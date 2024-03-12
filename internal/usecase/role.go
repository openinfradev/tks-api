package usecase

import (
	"github.com/openinfradev/tks-api/internal/model"
	"context"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IRoleUsecase interface {
	CreateTksRole(ctx context.Context, role *model.Role) (string, error)
	ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	GetTksRole(ctx context.Context, id string) (*model.Role, error)
	DeleteTksRole(ctx context.Context, id string) error
	UpdateTksRole(ctx context.Context, role *model.Role) error
}

type RoleUsecase struct {
	repo repository.IRoleRepository
}

func NewRoleUsecase(repo repository.Repository) *RoleUsecase {
	return &RoleUsecase{
		repo: repo.Role,
	}
}

func (r RoleUsecase) CreateTksRole(ctx context.Context, role *model.Role) (string, error) {
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

func (r RoleUsecase) DeleteTksRole(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}

func (r RoleUsecase) UpdateTksRole(ctx context.Context, role *model.Role) error {
	err := r.repo.Update(ctx, role)
	if err != nil {
		return err
	}

	return nil
}
