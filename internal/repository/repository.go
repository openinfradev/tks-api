package repository

import (
	"fmt"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/pagination"
)

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
}

func CombinedGormFilter(baseTableName string, filters []pagination.Filter) FilterFunc {
	return func(db *gorm.DB) *gorm.DB {
		for _, filter := range filters {
			if len(filter.Values) > 1 {
				inQuery := fmt.Sprintf("%s.%s in (", baseTableName, filter.Column)
				for _, val := range filter.Values {
					inQuery = inQuery + fmt.Sprintf("'%s',", val)
				}
				inQuery = inQuery[:len(inQuery)-1] + ")"
				db = db.Where(inQuery)
			} else {
				db = db.Where(fmt.Sprintf("%s.%s = '%s'", baseTableName, filter.Column, filter.Values[0]))
			}
		}
		return db
	}
}
