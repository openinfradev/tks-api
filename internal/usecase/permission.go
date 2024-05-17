package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
	"sort"
)

type IPermissionUsecase interface {
	CreatePermissionSet(ctx context.Context, permissionSet *model.PermissionSet) error
	GetPermissionSetByRoleId(ctx context.Context, roleId string) (*model.PermissionSet, error)
	ListPermissions(ctx context.Context, roleId string) ([]*model.Permission, error)
	SetRoleIdToPermissionSet(ctx context.Context, roleId string, permissionSet *model.PermissionSet)
	GetAllowedPermissionSet(ctx context.Context) *model.PermissionSet
	GetUserPermissionSet(ctx context.Context) *model.PermissionSet
	UpdatePermission(ctx context.Context, permission *model.Permission) error
	MergePermissionWithOrOperator(ctx context.Context, permissionSet ...*model.PermissionSet) *model.PermissionSet
	SyncKeycloakWithClusterAdminPermission(ctx context.Context, organizationId string, clientName string, userId string, roleName string, boolean bool) error
}

type PermissionUsecase struct {
	repo repository.IPermissionRepository
	kc   keycloak.IKeycloak
}

func NewPermissionUsecase(repo repository.Repository, kc keycloak.IKeycloak) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo.Permission,
		kc:   kc,
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
	if err = p.repo.Create(ctx, permissionSet.Policy); err != nil {
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
		Policy:            nil,
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
		case string(model.PolicyPermission):
			permissionSet.Policy = permission
		case string(model.ProjectPermission):
			permissionSet.ProjectManagement = permission
		case string(model.NotificationPermission):
			permissionSet.Notification = permission
		case string(model.ConfigurationPermission):
			permissionSet.Configuration = permission
		}

		p.sortPermissionRecursive(permission)
	}

	return permissionSet, nil
}

func (p PermissionUsecase) ListPermissions(ctx context.Context, roleId string) ([]*model.Permission, error) {
	permissions, err := p.repo.List(ctx, roleId)
	if err != nil {
		return nil, err
	}
	for _, permission := range permissions {
		p.sortPermissionRecursive(permission)
	}

	return permissions, nil
}

func (p PermissionUsecase) GetPermission(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	permission, err := p.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p.sortPermissionRecursive(permission)
	return permission, nil
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

func (p PermissionUsecase) MergePermissionWithOrOperator(ctx context.Context, permissionSet ...*model.PermissionSet) *model.PermissionSet {
	var out *model.PermissionSet
	for i, ps := range permissionSet {
		if i == 0 {
			out = ps
			continue
		}

		out.Dashboard = p.mergePermission(ctx, out.Dashboard, ps.Dashboard)
		out.Stack = p.mergePermission(ctx, out.Stack, ps.Stack)
		out.Policy = p.mergePermission(ctx, out.Policy, ps.Policy)
		out.ProjectManagement = p.mergePermission(ctx, out.ProjectManagement, ps.ProjectManagement)
		out.Notification = p.mergePermission(ctx, out.Notification, ps.Notification)
		out.Configuration = p.mergePermission(ctx, out.Configuration, ps.Configuration)
	}

	return out
}

func (p PermissionUsecase) mergePermission(ctx context.Context, mergedPermission, permission *model.Permission) *model.Permission {
	var mergedEdgePermissions []*model.Permission
	mergedEdgePermissions = model.GetEdgePermission(mergedPermission, mergedEdgePermissions, nil)

	var rightEdgePermissions []*model.Permission
	rightEdgePermissions = model.GetEdgePermission(permission, rightEdgePermissions, nil)

	for i, rightEdgePermission := range rightEdgePermissions {
		*(mergedEdgePermissions[i].IsAllowed) = *(mergedEdgePermissions[i].IsAllowed) || *(rightEdgePermission.IsAllowed)
	}

	return mergedPermission
}

func (p PermissionUsecase) sortPermissionRecursive(permission *model.Permission) {
	if len(permission.Children) > 0 {
		if permission.Children[0].IsAllowed != nil {
			sort.Slice(permission.Children, func(i, j int) bool {
				key1 := permission.Children[i].Key
				key2 := permission.Children[j].Key

				order1, exists1 := model.SortOrder[key1]
				order2, exists2 := model.SortOrder[key2]

				if exists1 && exists2 {
					return order1 < order2
				} else if exists1 {
					return true
				} else if exists2 {
					return false
				}

				return key1 < key2
			})
		}
		for _, child := range permission.Children {
			p.sortPermissionRecursive(child)
		}
	}
}

func (p PermissionUsecase) SyncKeycloakWithClusterAdminPermission(ctx context.Context, organizationId string, clientName string, userId string, roleName string, boolean bool) error {
	if boolean {
		return p.kc.AssignClientRoleToUser(ctx, organizationId, userId, clientName, roleName)
	} else {
		return p.kc.UnassignClientRoleToUser(ctx, organizationId, userId, clientName, roleName)
	}
}
