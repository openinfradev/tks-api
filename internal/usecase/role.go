package usecase

import (
	"context"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IRoleUsecase interface {
	CreateTksRole(ctx context.Context, role *model.Role) (string, error)
	ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	GetTksRole(ctx context.Context, orgainzationId string, id string) (*model.Role, error)
	DeleteTksRole(ctx context.Context, organizationId string, id string) error
	UpdateTksRole(ctx context.Context, role *model.Role) error
	IsRoleNameExisted(ctx context.Context, organizationId string, roleName string) (bool, error)
	SyncOldVersions(ctx context.Context) error
}

type RoleUsecase struct {
	repo           repository.IRoleRepository
	kc             keycloak.IKeycloak
	orgRepo        repository.IOrganizationRepository
	permissionRepo repository.IPermissionRepository
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

func (r RoleUsecase) GetTksRole(ctx context.Context, organizationId string, id string) (*model.Role, error) {
	role, err := r.repo.GetTksRole(ctx, organizationId, id)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r RoleUsecase) DeleteTksRole(ctx context.Context, organizationId string, id string) error {
	role, err := r.repo.GetTksRole(ctx, organizationId, id)
	if err != nil {
		return err
	}
	err = r.kc.DeleteGroup(ctx, organizationId, role.Name+"@"+organizationId)
	if err != nil {
		return err
	}
	return r.repo.Delete(ctx, id)
}

func (r RoleUsecase) UpdateTksRole(ctx context.Context, newRole *model.Role) error {
	role, err := r.repo.GetTksRole(ctx, newRole.OrganizationID, newRole.ID)
	if err != nil {
		return err
	}
	err = r.kc.UpdateGroup(ctx, role.OrganizationID, role.Name, newRole.Name)
	if err != nil {
		return err
	}
	err = r.repo.Update(ctx, newRole)
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

func (r RoleUsecase) SyncOldVersions(ctx context.Context) error {
	// Get all organizations
	orgs, _ := r.orgRepo.Fetch(ctx, nil)
	for _, org := range *orgs {
		roles, _ := r.repo.ListTksRoles(ctx, org.ID, nil)
		for _, role := range roles {
			storedPermissionSet := &model.PermissionSet{}

			permissionList, err := r.permissionRepo.List(ctx, role.ID)

			if err != nil {
				return err
			}
			for _, permission := range permissionList {
				switch permission.Name {
				case string(model.DashBoardPermission):
					storedPermissionSet.Dashboard = permission
					log.Debugf(ctx, "Dashboard Permission Set : %+v", storedPermissionSet.Dashboard)
				case string(model.StackPermission):
					storedPermissionSet.Stack = permission
					log.Debugf(ctx, "Stack Permission Set : %+v", storedPermissionSet.Stack)
				case string(model.PolicyPermission):
					storedPermissionSet.Policy = permission
					log.Debugf(ctx, "Policy Permission Set : %+v", storedPermissionSet.Policy)
				case string(model.ProjectPermission):
					storedPermissionSet.ProjectManagement = permission
					log.Debugf(ctx, "Project Permission Set : %+v", storedPermissionSet.ProjectManagement)
				case string(model.NotificationPermission):
					storedPermissionSet.Notification = permission
					log.Debugf(ctx, "Notification Permission Set : %+v", storedPermissionSet.Notification)
				case string(model.ConfigurationPermission):
					storedPermissionSet.Configuration = permission
					log.Debugf(ctx, "Configuration Permission Set : %+v", storedPermissionSet.Configuration)
				}
			}

			// tmp
			t := model.NewDefaultPermissionSet()
			var overwritePermissions []*model.Permission
			overwritePermissions = make([]*model.Permission, 0)

			// dashboard
			storedPermissionSet.Dashboard.Children[0].Children[0].EdgeKey = t.Dashboard.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Dashboard.Children[0].Children[0])
			storedPermissionSet.Dashboard.Children[0].Children[1].EdgeKey = t.Dashboard.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Dashboard.Children[0].Children[1])

			// stack
			storedPermissionSet.Stack.Children[0].Children[0].EdgeKey = t.Stack.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Stack.Children[0].Children[0])
			storedPermissionSet.Stack.Children[0].Children[1].EdgeKey = t.Stack.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Stack.Children[0].Children[1])
			storedPermissionSet.Stack.Children[0].Children[2].EdgeKey = t.Stack.Children[0].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Stack.Children[0].Children[2])
			storedPermissionSet.Stack.Children[0].Children[3].EdgeKey = t.Stack.Children[0].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Stack.Children[0].Children[3])

			// policy
			storedPermissionSet.Policy.Children[0].Children[0].EdgeKey = t.Policy.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Policy.Children[0].Children[0])
			storedPermissionSet.Policy.Children[0].Children[1].EdgeKey = t.Policy.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Policy.Children[0].Children[1])
			storedPermissionSet.Policy.Children[0].Children[2].EdgeKey = t.Policy.Children[0].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Policy.Children[0].Children[2])
			storedPermissionSet.Policy.Children[0].Children[3].EdgeKey = t.Policy.Children[0].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Policy.Children[0].Children[3])

			// notification
			storedPermissionSet.Notification.Children[0].Children[0].EdgeKey = t.Notification.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Notification.Children[0].Children[0])
			storedPermissionSet.Notification.Children[0].Children[1].EdgeKey = t.Notification.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Notification.Children[0].Children[1])
			storedPermissionSet.Notification.Children[0].Children[2].EdgeKey = t.Notification.Children[0].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Notification.Children[0].Children[2])
			storedPermissionSet.Notification.Children[1].Children[0].EdgeKey = t.Notification.Children[1].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Notification.Children[1].Children[0])
			storedPermissionSet.Notification.Children[1].Children[1].EdgeKey = t.Notification.Children[1].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Notification.Children[1].Children[1])

			// project
			// 1depth
			storedPermissionSet.ProjectManagement.Children[0].Children[0].EdgeKey = t.ProjectManagement.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[0].Children[0])
			storedPermissionSet.ProjectManagement.Children[0].Children[1].EdgeKey = t.ProjectManagement.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[0].Children[1])
			storedPermissionSet.ProjectManagement.Children[0].Children[2].EdgeKey = t.ProjectManagement.Children[0].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[0].Children[2])
			storedPermissionSet.ProjectManagement.Children[0].Children[3].EdgeKey = t.ProjectManagement.Children[0].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[0].Children[3])

			// 2depth
			storedPermissionSet.ProjectManagement.Children[1].Children[0].EdgeKey = t.ProjectManagement.Children[1].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[1].Children[0])
			storedPermissionSet.ProjectManagement.Children[1].Children[1].EdgeKey = t.ProjectManagement.Children[1].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[1].Children[1])

			// 3depth
			storedPermissionSet.ProjectManagement.Children[2].Children[0].EdgeKey = t.ProjectManagement.Children[2].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[2].Children[0])
			storedPermissionSet.ProjectManagement.Children[2].Children[1].EdgeKey = t.ProjectManagement.Children[2].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[2].Children[1])
			storedPermissionSet.ProjectManagement.Children[2].Children[2].EdgeKey = t.ProjectManagement.Children[2].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[2].Children[2])
			storedPermissionSet.ProjectManagement.Children[2].Children[3].EdgeKey = t.ProjectManagement.Children[2].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[2].Children[3])

			// 4depth
			storedPermissionSet.ProjectManagement.Children[3].Children[0].EdgeKey = t.ProjectManagement.Children[3].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[3].Children[0])
			storedPermissionSet.ProjectManagement.Children[3].Children[1].EdgeKey = t.ProjectManagement.Children[3].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[3].Children[1])
			storedPermissionSet.ProjectManagement.Children[3].Children[2].EdgeKey = t.ProjectManagement.Children[3].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[3].Children[2])
			storedPermissionSet.ProjectManagement.Children[3].Children[3].EdgeKey = t.ProjectManagement.Children[3].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[3].Children[3])

			// 5depth
			storedPermissionSet.ProjectManagement.Children[4].Children[0].EdgeKey = t.ProjectManagement.Children[4].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[4].Children[0])
			storedPermissionSet.ProjectManagement.Children[4].Children[1].EdgeKey = t.ProjectManagement.Children[4].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[4].Children[1])
			storedPermissionSet.ProjectManagement.Children[4].Children[2].EdgeKey = t.ProjectManagement.Children[4].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[4].Children[2])
			storedPermissionSet.ProjectManagement.Children[4].Children[3].EdgeKey = t.ProjectManagement.Children[4].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.ProjectManagement.Children[4].Children[3])

			// configuration
			storedPermissionSet.Configuration.Children[0].Children[0].EdgeKey = t.Configuration.Children[0].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[0].Children[0])
			storedPermissionSet.Configuration.Children[0].Children[1].EdgeKey = t.Configuration.Children[0].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[0].Children[1])

			// 2depth
			storedPermissionSet.Configuration.Children[1].Children[0].EdgeKey = t.Configuration.Children[1].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[1].Children[0])
			storedPermissionSet.Configuration.Children[1].Children[1].EdgeKey = t.Configuration.Children[1].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[1].Children[1])
			storedPermissionSet.Configuration.Children[1].Children[2].EdgeKey = t.Configuration.Children[1].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[1].Children[2])
			storedPermissionSet.Configuration.Children[1].Children[3].EdgeKey = t.Configuration.Children[1].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[1].Children[3])

			// 3depth
			storedPermissionSet.Configuration.Children[2].Children[0].EdgeKey = t.Configuration.Children[2].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[2].Children[0])
			storedPermissionSet.Configuration.Children[2].Children[1].EdgeKey = t.Configuration.Children[2].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[2].Children[1])

			// 4depth
			storedPermissionSet.Configuration.Children[3].Children[0].EdgeKey = t.Configuration.Children[3].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[3].Children[0])
			storedPermissionSet.Configuration.Children[3].Children[1].EdgeKey = t.Configuration.Children[3].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[3].Children[1])
			storedPermissionSet.Configuration.Children[3].Children[2].EdgeKey = t.Configuration.Children[3].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[3].Children[2])
			storedPermissionSet.Configuration.Children[3].Children[3].EdgeKey = t.Configuration.Children[3].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[3].Children[3])

			// 5depth
			storedPermissionSet.Configuration.Children[4].Children[0].EdgeKey = t.Configuration.Children[4].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[4].Children[0])
			storedPermissionSet.Configuration.Children[4].Children[1].EdgeKey = t.Configuration.Children[4].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[4].Children[1])
			storedPermissionSet.Configuration.Children[4].Children[2].EdgeKey = t.Configuration.Children[4].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[4].Children[2])
			storedPermissionSet.Configuration.Children[4].Children[3].EdgeKey = t.Configuration.Children[4].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[4].Children[3])

			// 6depth
			storedPermissionSet.Configuration.Children[5].Children[0].EdgeKey = t.Configuration.Children[5].Children[0].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[5].Children[0])
			storedPermissionSet.Configuration.Children[5].Children[1].EdgeKey = t.Configuration.Children[5].Children[1].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[5].Children[1])
			storedPermissionSet.Configuration.Children[5].Children[2].EdgeKey = t.Configuration.Children[5].Children[2].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[5].Children[2])
			storedPermissionSet.Configuration.Children[5].Children[3].EdgeKey = t.Configuration.Children[5].Children[3].EdgeKey
			overwritePermissions = append(overwritePermissions, storedPermissionSet.Configuration.Children[5].Children[3])

			for _, permission := range overwritePermissions {
				if err = r.permissionRepo.EdgeKeyOverwrite(ctx, permission); err != nil {
					return err
				}
			}
			log.Debugf(ctx, "Dashboard EdgeKey Overwrite Success")
		}
	}

	return nil
}
