package role

import internalApi "github.com/openinfradev/tks-api/internal/delivery/api"

type Role string

const (
	Admin  Role = "admin"
	User   Role = "user"
	leader Role = "leader"
	member Role = "member"
	viewer Role = "viewer"
)

func init() {
	RBAC = make(map[internalApi.Endpoint]map[Role]struct{})

	defaultPermissions := getDefaultPermissions()
	registerRole(defaultPermissions...)
}

func (r Role) String() string {
	return string(r)
}

func registerRole(dps ...*defaultPermission) {
	for _, dp := range dps {
		for _, endpoint := range *dp.permissions {
			if RBAC[endpoint] == nil {
				RBAC[endpoint] = make(map[Role]struct{})
			}
			RBAC[endpoint][dp.role] = struct{}{}
		}
	}
}

var RBAC map[internalApi.Endpoint]map[Role]struct{}

func StrToRole(role string) Role {
	switch role {
	case "admin":
		return Admin
	case "user":
		return User
	case "leader":
		return leader
	default:
		return ""
	}
}
func IsRoleAllowed(endpoint internalApi.Endpoint, role Role) bool {
	if RBAC[endpoint] == nil {
		return false
	}

	if _, ok := RBAC[endpoint][role]; ok {
		return true
	}

	return false
}
