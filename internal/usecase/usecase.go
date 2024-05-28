package usecase

import "github.com/openinfradev/tks-api/internal/repository"

type Usecase struct {
	Auth                       IAuthUsecase
	User                       IUserUsecase
	Cluster                    IClusterUsecase
	Organization               IOrganizationUsecase
	AppGroup                   IAppGroupUsecase
	AppServeApp                IAppServeAppUsecase
	CloudAccount               ICloudAccountUsecase
	StackTemplate              IStackTemplateUsecase
	Dashboard                  IDashboardUsecase
	SystemNotification         ISystemNotificationUsecase
	SystemNotificationTemplate ISystemNotificationTemplateUsecase
	SystemNotificationRule     ISystemNotificationRuleUsecase
	Stack                      IStackUsecase
	Project                    IProjectUsecase
	Role                       IRoleUsecase
	Permission                 IPermissionUsecase
	Audit                      IAuditUsecase
	PolicyTemplate             IPolicyTemplateUsecase
	Policy                     IPolicyUsecase
}

func NewUsecaseFactory(rf repository.Repository) *Usecase {
	return &Usecase{
		StackTemplate: NewStackTemplateUsecase(rf),
	}

}
