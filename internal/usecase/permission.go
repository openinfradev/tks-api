package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IPermissionUsecase interface {
	CreatePermissionSet(permissionSet *domain.PermissionSet) error
	GetPermissionSetByRoleId(roleId string) (*domain.PermissionSet, error)
	ListPermissions(roleId string) ([]*domain.Permission, error)
	//GetPermission(id uuid.UUID) (*domain.Permission, error)
	//DeletePermission(id uuid.UUID) error
	//UpdatePermission(permission *domain.Permission) error
	SetRoleIdToPermissionSet(roleId string, permissionSet *domain.PermissionSet)
	GetAllowedPermissionSet() *domain.PermissionSet
	GetUserPermissionSet() *domain.PermissionSet
	UpdatePermission(permission *domain.Permission) error
}

type PermissionUsecase struct {
	repo repository.IPermissionRepository
}

func NewPermissionUsecase(repo repository.Repository) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo.Permission,
	}
}

func (p PermissionUsecase) CreatePermissionSet(permissionSet *domain.PermissionSet) error {
	var err error
	if err = p.repo.Create(permissionSet.Dashboard); err != nil {
		return err
	}
	if err = p.repo.Create(permissionSet.Stack); err != nil {
		return err
	}
	if err = p.repo.Create(permissionSet.SecurityPolicy); err != nil {
		return err
	}
	if err = p.repo.Create(permissionSet.ProjectManagement); err != nil {
		return err
	}
	if err = p.repo.Create(permissionSet.Notification); err != nil {
		return err
	}
	if err = p.repo.Create(permissionSet.Configuration); err != nil {
		return err
	}

	return nil
}
func (p PermissionUsecase) GetPermissionSetByRoleId(roleId string) (*domain.PermissionSet, error) {
	permissionSet := &domain.PermissionSet{
		Dashboard:         nil,
		Stack:             nil,
		SecurityPolicy:    nil,
		ProjectManagement: nil,
		Notification:      nil,
		Configuration:     nil,
	}

	permissionList, err := p.repo.List(roleId)
	if err != nil {
		return nil, err
	}
	for _, permission := range permissionList {
		switch permission.Name {
		case string(domain.DashBoardPermission):
			permissionSet.Dashboard = permission
		case string(domain.StackPermission):
			permissionSet.Stack = permission
		case string(domain.SecurityPolicyPermission):
			permissionSet.SecurityPolicy = permission
		case string(domain.ProjectManagementPermission):
			permissionSet.ProjectManagement = permission
		case string(domain.NotificationPermission):
			permissionSet.Notification = permission
		case string(domain.ConfigurationPermission):
			permissionSet.Configuration = permission
		}
	}

	return permissionSet, nil
}

func (p PermissionUsecase) ListPermissions(roleId string) ([]*domain.Permission, error) {
	return p.repo.List(roleId)
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

func (p PermissionUsecase) SetRoleIdToPermissionSet(roleId string, permissionSet *domain.PermissionSet) {
	permissionSet.SetRoleId(roleId)
}

func (p PermissionUsecase) GetAllowedPermissionSet() *domain.PermissionSet {
	permissionSet := domain.NewDefaultPermissionSet()
	permissionSet.SetAllowedPermissionSet()
	return permissionSet
}

func (p PermissionUsecase) GetUserPermissionSet() *domain.PermissionSet {
	permissionSet := domain.NewDefaultPermissionSet()
	permissionSet.SetUserPermissionSet()
	return permissionSet
}
