package repository

import "gorm.io/gorm"

type FilterFunc func(user *gorm.DB) *gorm.DB

type Repository struct {
	User          IUserRepository
	Cluster       IClusterRepository
	Organization  IOrganizationRepository
	AppGroup      IAppGroupRepository
	AppServeApp   IAppServeAppRepository
	CloudSetting  ICloudSettingRepository
	StackTemplate IStackTemplateRepository
	History       IHistoryRepository
}
