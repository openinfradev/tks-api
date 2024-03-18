package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IPermissionUsecase interface {
	CreatePermissionSet(ctx context.Context, permissionSet *model.PermissionSet) error
	GetPermissionSetByRoleId(ctx context.Context, roleId string) (*model.PermissionSet, error)
	ListPermissions(ctx context.Context, roleId string) ([]*model.Permission, error)
	SetRoleIdToPermissionSet(ctx context.Context, roleId string, permissionSet *model.PermissionSet)
	GetAllowedPermissionSet(ctx context.Context) *model.PermissionSet
	GetUserPermissionSet(ctx context.Context) *model.PermissionSet
	UpdatePermission(ctx context.Context, permission *model.Permission) error
}

type PermissionUsecase struct {
	repo repository.IPermissionRepository
}

func NewPermissionUsecase(repo repository.Repository) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo.Permission,
	}
}

func (p PermissionUsecase) CreatePermissionSet(ctx context.Context, permissionSet *model.PermissionSet) error {
	var err error
	if err = p.repo.Create(ctx, permissionSet.Dashboard); err != nil {
		return err
	}
	if err = p.repo.Create(ctx, permissionSet.Stack); err != nil {
		return err
	}
	if err = p.repo.Create(ctx, permissionSet.SecurityPolicy); err != nil {
		return err
	}
	if err = p.repo.Create(ctx, permissionSet.ProjectManagement); err != nil {
		return err
	}
	if err = p.repo.Create(ctx, permissionSet.Notification); err != nil {
		return err
	}
	if err = p.repo.Create(ctx, permissionSet.Configuration); err != nil {
		return err
	}

	return nil
}
func (p PermissionUsecase) GetPermissionSetByRoleId(ctx context.Context, roleId string) (*model.PermissionSet, error) {
	permissionSet := &model.PermissionSet{
		Dashboard:         nil,
		Stack:             nil,
		SecurityPolicy:    nil,
		ProjectManagement: nil,
		Notification:      nil,
		Configuration:     nil,
	}

	permissionList, err := p.repo.List(ctx, roleId)
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

func (p PermissionUsecase) ListPermissions(ctx context.Context, roleId string) ([]*model.Permission, error) {
	return p.repo.List(ctx, roleId)
}

func (p PermissionUsecase) GetPermission(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	return p.repo.Get(ctx, id)
}

func (p PermissionUsecase) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return p.repo.Delete(ctx, id)
}

func (p PermissionUsecase) UpdatePermission(ctx context.Context, permission *model.Permission) error {
	return p.repo.Update(ctx, permission)
}

func (p PermissionUsecase) SetRoleIdToPermissionSet(ctx context.Context, roleId string, permissionSet *model.PermissionSet) {
	permissionSet.SetRoleId(roleId)
}

func (p PermissionUsecase) GetAllowedPermissionSet(ctx context.Context) *model.PermissionSet {
	permissionSet := model.NewDefaultPermissionSet()
	permissionSet.SetAllowedPermissionSet()
	return permissionSet
}

func (p PermissionUsecase) GetUserPermissionSet(ctx context.Context) *model.PermissionSet {
	permissionSet := model.NewDefaultPermissionSet()
	permissionSet.SetUserPermissionSet()
	return permissionSet
}
