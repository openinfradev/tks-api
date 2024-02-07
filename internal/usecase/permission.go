package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IPermissionUsecase interface {
	CreatePermission(permission *domain.Permission) error
	ListPermissions() ([]*domain.Permission, error)
	GetPermission(id uuid.UUID) (*domain.Permission, error)
	DeletePermission(id uuid.UUID) error
	UpdatePermission(permission *domain.Permission) error
}

type PermissionUsecase struct {
	repo repository.IPermissionRepository
}

func NewPermissionUsecase(repo repository.IPermissionRepository) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo,
	}
}

func (p PermissionUsecase) CreatePermission(permission *domain.Permission) error {
	return p.repo.Create(permission)
}

func (p PermissionUsecase) ListPermissions() ([]*domain.Permission, error) {
	return p.repo.List()
}

func (p PermissionUsecase) GetPermission(id uuid.UUID) (*domain.Permission, error) {
	return p.repo.Get(id)
}

func (p PermissionUsecase) DeletePermission(id uuid.UUID) error {
	return p.repo.Delete(id)
}

func (p PermissionUsecase) UpdatePermission(permission *domain.Permission) error {
	return p.repo.Update(permission)
}
