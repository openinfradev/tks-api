package repository

import "gorm.io/gorm"

type FilterFunc func(user *gorm.DB) *gorm.DB

type Repository struct {
	Auth          IAuthRepository
	User          IUserRepository
	Cluster       IClusterRepository
	Organization  IOrganizationRepository
	AppGroup      IAppGroupRepository
	AppServeApp   IAppServeAppRepository
	CloudAccount  ICloudAccountRepository
	StackTemplate IStackTemplateRepository
	Alert         IAlertRepository
	History       IHistoryRepository
}
