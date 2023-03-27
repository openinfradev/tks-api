package user

import "github.com/google/uuid"

// Info describes a user that has been authenticated to the system.
type Info interface {
	GetUserId() uuid.UUID
	GetOrganizationId() string
	GetRoleProjectMapping() map[string]string
}

// DefaultInfo provides a simple user information exchange object
// for components that implement the UserInfo interface.
type DefaultInfo struct {
	UserId             uuid.UUID
	OrganizationId     string
	RoleProjectMapping map[string]string
}

func (i *DefaultInfo) GetUserId() uuid.UUID {
	return i.UserId
}

func (i *DefaultInfo) GetOrganizationId() string {
	return i.OrganizationId
}

// GetRoleGroupMapping key is project name, value is role name
func (i *DefaultInfo) GetRoleProjectMapping() map[string]string {
	return i.RoleProjectMapping
}

// well-known user and group names
const (
	TksAdminRole  = "tks_admin"
	AdminRole     = "admin"
	ProjectLeader = "project_leader"
	ProjectMember = "project_member"
	ProjectViewer = "project_viewer"
)
