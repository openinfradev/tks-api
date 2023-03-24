package keycloak

type Config struct {
	Address       string
	ClientSecret  string
	AdminId       string
	AdminPassword string
}

const (
	DefaultMasterRealm  = "master"
	DefaultClientID     = "tks"
	DefaultClientSecret = "secret"
)
