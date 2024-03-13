package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IPermissionUsecase interface {
	CreatePermissionSet(permissionSet *model.PermissionSet) error
	GetPermissionSetByRoleId(roleId string) (*model.PermissionSet, error)
	ListPermissions(roleId string) ([]*model.Permission, error)
	//GetPermission(id uuid.UUID) (*model.Permission, error)
	//DeletePermission(id uuid.UUID) error
	//UpdatePermission(permission *model.Permission) error
	SetRoleIdToPermissionSet(roleId string, permissionSet *model.PermissionSet)
	GetAllowedPermissionSet() *model.PermissionSet
	GetUserPermissionSet() *model.PermissionSet
	UpdatePermission(permission *model.Permission) error
}

type PermissionUsecase struct {
	repo repository.IPermissionRepository
}

func NewPermissionUsecase(repo repository.Repository) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo.Permission,
	}
}

func (p PermissionUsecase) CreatePermissionSet(permissionSet *model.PermissionSet) error {
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
func (p PermissionUsecase) GetPermissionSetByRoleId(roleId string) (*model.PermissionSet, error) {
	permissionSet := &model.PermissionSet{
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
		case string(model.DashBoardPermission):
			permissionSet.Dashboard = permission
		case string(model.StackPermission):
			permissionSet.Stack = permission
		case string(model.SecurityPolicyPermission):
			permissionSet.SecurityPolicy = permission
		case string(model.ProjectManagementPermission):
			permissionSet.ProjectManagement = permission
		case string(model.NotificationPermission):
			permissionSet.Notification = permission
		case string(model.ConfigurationPermission):
			permissionSet.Configuration = permission
		}
	}

	return permissionSet, nil
}

func (p PermissionUsecase) ListPermissions(roleId string) ([]*model.Permission, error) {
	return p.repo.List(roleId)
}

func (p PermissionUsecase) GetPermission(id uuid.UUID) (*model.Permission, error) {
	return p.repo.Get(id)
}

func (p PermissionUsecase) DeletePermission(id uuid.UUID) error {
	return p.repo.Delete(id)
}

func (p PermissionUsecase) UpdatePermission(permission *model.Permission) error {
	return p.repo.Update(permission)
}

func (p PermissionUsecase) SetRoleIdToPermissionSet(roleId string, permissionSet *model.PermissionSet) {
	permissionSet.SetRoleId(roleId)
}

func (p PermissionUsecase) GetAllowedPermissionSet() *model.PermissionSet {
	permissionSet := model.NewDefaultPermissionSet()
	permissionSet.SetAllowedPermissionSet()
	return permissionSet
}

func (p PermissionUsecase) GetUserPermissionSet() *model.PermissionSet {
	permissionSet := model.NewDefaultPermissionSet()
	permissionSet.SetUserPermissionSet()
	return permissionSet
}
