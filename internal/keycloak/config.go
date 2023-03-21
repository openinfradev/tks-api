package keycloak

type Config struct {
	Address       string
	ClientSecret  string
	AdminId       string
	AdminPassword string
	MasterRealm   string
}

const (
	DefaultMasterRealm  = "master"
	DefaultClientID     = "tks"
	DefaultClientSecret = "secret"
)
