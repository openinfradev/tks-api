package usecase

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
	Alert                      IAlertUsecase
	SystemNotificationTemplate ISystemNotificationTemplateUsecase
	Stack                      IStackUsecase
	Project                    IProjectUsecase
	Role                       IRoleUsecase
	Permission                 IPermissionUsecase
	Audit                      IAuditUsecase
	PolicyTemplate             IPolicyTemplateUsecase
}
