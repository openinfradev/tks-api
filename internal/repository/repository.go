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

func CombinedGormFilter(filters []pagination.Filter, combinedFilter pagination.CombinedFilter) FilterFunc {
	return func(db *gorm.DB) *gorm.DB {
		// and query
		for _, filter := range filters {
			if len(filter.Values) > 1 {
				inQuery := fmt.Sprintf("%s in (", filter.Column)
				for _, val := range filter.Values {
					inQuery = inQuery + fmt.Sprintf("'%s',", val)
				}
				inQuery = inQuery[:len(inQuery)-1] + ")"
				db = db.Where(inQuery)
			} else {
				db = db.Where(fmt.Sprintf("%s = '%s'", filter.Column, filter.Values[0]))
			}
		}

		// or query
		// id = '123' or description = '345'
		if len(combinedFilter.Columns) > 0 {
			orQuery := ""
			for _, column := range combinedFilter.Columns {
				orQuery = orQuery + fmt.Sprintf("%s like '%%%s%%' OR ", column, combinedFilter.Value)
			}
			orQuery = orQuery[:len(orQuery)-3]
			db = db.Where(orQuery)
		}

		return db
	}
}
