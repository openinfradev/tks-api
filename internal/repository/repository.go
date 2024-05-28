package repository

import (
	"gorm.io/gorm"
)

type FilterFunc func(user *gorm.DB) *gorm.DB

type Repository struct {
	Auth                       IAuthRepository
	User                       IUserRepository
	Cluster                    IClusterRepository
	Organization               IOrganizationRepository
	AppGroup                   IAppGroupRepository
	AppServeApp                IAppServeAppRepository
	CloudAccount               ICloudAccountRepository
	StackTemplate              IStackTemplateRepository
	Role                       IRoleRepository
	Permission                 IPermissionRepository
	Endpoint                   IEndpointRepository
	Project                    IProjectRepository
	Audit                      IAuditRepository
	PolicyTemplate             IPolicyTemplateRepository
	Policy                     IPolicyRepository
	SystemNotification         ISystemNotificationRepository
	SystemNotificationTemplate ISystemNotificationTemplateRepository
	SystemNotificationRule     ISystemNotificationRuleRepository
	Dashboard                  IDashboardRepository
}

func NewRepositoryFactory(db *gorm.DB) *Repository {
	return &Repository{
		Auth:                       NewAuthRepository(db),
		User:                       NewUserRepository(db),
		Cluster:                    NewClusterRepository(db),
		Organization:               NewOrganizationRepository(db),
		AppGroup:                   NewAppGroupRepository(db),
		AppServeApp:                NewAppServeAppRepository(db),
		CloudAccount:               NewCloudAccountRepository(db),
		StackTemplate:              NewStackTemplateRepository(db),
		SystemNotification:         NewSystemNotificationRepository(db),
		SystemNotificationTemplate: NewSystemNotificationTemplateRepository(db),
		SystemNotificationRule:     NewSystemNotificationRuleRepository(db),
		Role:                       NewRoleRepository(db),
		Project:                    NewProjectRepository(db),
		Permission:                 NewPermissionRepository(db),
		Endpoint:                   NewEndpointRepository(db),
		Audit:                      NewAuditRepository(db),
		PolicyTemplate:             NewPolicyTemplateRepository(db),
		Policy:                     NewPolicyRepository(db),
		Dashboard:                  NewDashboardRepository(db),
	}
}
