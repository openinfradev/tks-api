package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/pagination"
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
	Alert                      IAlertRepository
	Role                       IRoleRepository
	Permission                 IPermissionRepository
	Endpoint                   IEndpointRepository
	Project                    IProjectRepository
	Audit                      IAuditRepository
	PolicyTemplate             IPolicyTemplateRepository
	SystemNotificationTemplate ISystemNotificationTemplateRepository
}

func CombinedGormFilter(table string, filters []pagination.Filter, combinedFilter pagination.CombinedFilter) FilterFunc {
	return func(db *gorm.DB) *gorm.DB {
		// and query
		for _, filter := range filters {
			if len(filter.Values) > 1 {
				inQuery := fmt.Sprintf("LOWER(%s.%s::text) in (", table, filter.Column)
				for _, val := range filter.Values {
					inQuery = inQuery + fmt.Sprintf("LOWER($$%s$$),", val)
				}
				inQuery = inQuery[:len(inQuery)-1] + ")"
				db = db.Where(inQuery)
			} else {
				if len(filter.Values[0]) > 0 {
					if strings.Contains(filter.Values[0], "%") {
						filterVal := strings.Replace(filter.Values[0], "%", "`%", -1)
						db = db.Where(fmt.Sprintf("LOWER(%s.%s::text) like LOWER($$%%%s%%$$) escape '`'", table, filter.Column, filterVal))
					} else {
						db = db.Where(fmt.Sprintf("LOWER(%s.%s::text) like LOWER($$%%%s%%$$)", table, filter.Column, filter.Values[0]))
					}
				}
			}
		}

		// or query
		// id = '123' or description = '345'
		if len(combinedFilter.Columns) > 0 {
			orQuery := ""
			for _, column := range combinedFilter.Columns {
				if strings.Contains(combinedFilter.Value, "%") {
					filterVal := strings.Replace(combinedFilter.Value, "%", "`%", -1)
					orQuery = orQuery + fmt.Sprintf("LOWER(%s.%s::text) like LOWER($$%%%s%%$$) escape '`' OR ", table, column, filterVal)
				} else {
					orQuery = orQuery + fmt.Sprintf("LOWER(%s.%s::text) like LOWER($$%%%s%%$$) OR ", table, column, combinedFilter.Value)
				}
			}
			orQuery = orQuery[:len(orQuery)-3]
			db = db.Where(orQuery)
		}

		return db
	}
}
