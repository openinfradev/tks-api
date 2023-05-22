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
	accessTokenLifespan   = 60 * 60 * 24 // 1 day
	ssoSessionIdleTimeout = 60 * 60 * 8  // 8 hours
	ssoSessionMaxLifespan = 60 * 60 * 24 // 1 day
)
