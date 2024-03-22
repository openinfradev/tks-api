package keycloak

type Config struct {
	Address       string
	ClientSecret  string
	AdminId       string
	AdminPassword string
}

const (
	DefaultMasterRealm    = "master"
	DefaultClientID       = "tks"
	DefaultClientSecret   = "secret"
	AdminCliClientID      = "admin-cli"
	AccessTokenLifespan   = 60 * 60 * 24 // 1 day
	SsoSessionIdleTimeout = 60 * 60 * 24 // 1 day
	SsoSessionMaxLifespan = 60 * 60 * 24 // 1 day
)
