package role

import internalApi "github.com/openinfradev/tks-api/internal/delivery/api"

type Role string

const (
	Admin  Role = "admin"
	User   Role = "user"
	leader Role = "leader"
)

func (r Role) String() string {
	return string(r)
}

var RBAC = map[internalApi.Endpoint][]Role{
	internalApi.CreateProject: {Admin, User, leader},
	internalApi.UpdateProject: {Admin, User, leader},
	internalApi.DeleteProject: {Admin, User, leader},
	internalApi.GetProject:    {Admin, User, leader},
	internalApi.GetProjects:   {Admin, User, leader},
}

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
	for _, r := range RBAC[endpoint] {
		if r == role {
			return true
		}
	}
	return false
}
