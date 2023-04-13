package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/openinfradev/tks-api/pkg/httpErrors"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v4"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IKeycloak interface {
	InitializeKeycloak() error

	LoginAdmin() (string, error)

	CreateRealm(organizationName string, organizationConfig domain.Organization, token string) (string, error)
	GetRealm(organizationName string, token string) (*domain.Organization, error)
	GetRealms(token string) ([]*domain.Organization, error)
	DeleteRealm(organizationName string, token string) error
	UpdateRealm(organizationName string, organizationConfig domain.Organization, token string) error

	CreateUser(organizationName string, user *gocloak.User, token string) error
	GetUser(organizationName string, userAccountId string, token string) (*gocloak.User, error)
	GetUsers(organizationName string, token string) ([]*gocloak.User, error)
	DeleteUser(organizationName string, userAccountId string, token string) error
	UpdateUser(organizationName string, user *gocloak.User, accessToken string) error

	GetAccessTokenByIdPassword(accountId string, password string, organizationName string) (*domain.User, error)
	VerifyAccessToken(token string, organizationName string) error
	ParseAccessToken(token string, organization string) (*jwt.Token, *jwt.MapClaims, error)
}

type Keycloak struct {
	config *Config
	client *gocloak.GoCloak
}

func (c *Keycloak) LoginAdmin() (string, error) {
	token, err := c.client.LoginAdmin(context.Background(), c.config.AdminId, c.config.AdminPassword, DefaultMasterRealm)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func New(config *Config) IKeycloak {
	return &Keycloak{
		config: config,
	}
}
func (c *Keycloak) InitializeKeycloak() error {
	c.client = gocloak.NewClient(c.config.Address)
	ctx := context.Background()
	restyClient := c.client.RestyClient()
	//for debugging

	//if os.Getenv("LOG_LEVEL") == "DEBUG" {
	//	restyClient.SetDebug(true)
	//}

	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	token, err := c.loginAdmin(ctx)
	if err != nil {
		log.Fatal(err)
		return err
	}

	group, err := c.ensureGroupByName(ctx, token, DefaultMasterRealm, "tks-admin@master")
	if err != nil {
		log.Fatal(err)
		return err
	}

	user, err := c.ensureUserByName(ctx, token, DefaultMasterRealm, c.config.AdminId, c.config.AdminPassword)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err := c.addUserToGroup(ctx, token, DefaultMasterRealm, *user.ID, *group.ID); err != nil {
		log.Fatal(err)
		return err
	}

	keycloakClient, err := c.ensureClient(ctx, token, DefaultMasterRealm, DefaultClientID, DefaultClientSecret)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, defaultMapper := range defaultProtocolTksMapper {
		if err := c.ensureClientProtocolMappers(ctx, token, DefaultMasterRealm, *keycloakClient.ClientID, "openid", defaultMapper); err != nil {
			log.Fatal(err)
			return err
		}
	}

	if _, err := c.client.Login(ctx, DefaultClientID, DefaultClientSecret, DefaultMasterRealm,
		c.config.AdminId, c.config.AdminPassword); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (k *Keycloak) CreateRealm(organizationName string, organizationConfig domain.Organization, accessToken string) (string, error) {
	//TODO implement me
	ctx := context.Background()

	realmConfig := gocloak.RealmRepresentation{
		Realm:               &organizationName,
		Enabled:             gocloak.BoolP(true),
		AccessTokenLifespan: gocloak.IntP(60 * 60 * 24),
	}
	realmUUID, err := k.client.CreateRealm(ctx, accessToken, realmConfig)
	if err != nil {
		return realmUUID, err
	}
	// After Create Realm, accesstoken got changed so that old token doesn't work properly.
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return realmUUID, err
	}
	accessToken = token.AccessToken

	time.Sleep(time.Second * 3)
	clientUUID, err := k.createDefaultClient(context.Background(), accessToken, organizationName, DefaultClientID, DefaultClientSecret)
	if err != nil {
		log.Error(err, "createDefaultClient")
		return realmUUID, err
	}

	token, err = k.loginAdmin(ctx)
	if err != nil {
		return realmUUID, err
	}
	accessToken = token.AccessToken

	for _, defaultMapper := range defaultProtocolTksMapper {
		if *defaultMapper.Name == "org" {
			defaultMapper.Config = &map[string]string{
				"full.path":            "false",
				"id.token.claim":       "false",
				"access.token.claim":   "true",
				"claim.name":           "organization",
				"claim.value":          organizationName,
				"userinfo.token.claim": "false",
			}
		}
		if _, err := k.createClientProtocolMapper(ctx, accessToken, organizationName, clientUUID, defaultMapper); err != nil {
			return realmUUID, err
		}
	}
	adminGroupUuid, err := k.createGroup(ctx, accessToken, organizationName, "admin@"+organizationName)
	if err != nil {
		return realmUUID, err
	}

	token, err = k.loginAdmin(ctx)
	if err != nil {
		return realmUUID, err
	}
	accessToken = token.AccessToken

	realmManagementClientUuid, err := k.getClientByClientId(ctx, accessToken, organizationName, "realm-management")
	if err != nil {
		return realmUUID, err
	}

	//token, err = k.loginAdmin(ctx)
	//if err != nil {
	//	return realmUUID, err
	//}
	realmAdminRole, err := k.getClientRole(ctx, accessToken, organizationName, realmManagementClientUuid, "realm-admin")
	if err != nil {
		return realmUUID, err
	}

	err = k.addClientRoleToGroup(ctx, accessToken, organizationName, realmManagementClientUuid, adminGroupUuid,
		&gocloak.Role{
			ID:   realmAdminRole.ID,
			Name: realmAdminRole.Name,
		})

	if err != nil {
		return realmUUID, err
	}

	userGroupUuid, err := k.createGroup(ctx, accessToken, organizationName, "user@"+organizationName)
	if err != nil {
		return realmUUID, err
	}
	//adminGroup, err := k.ensureGroup(ctx, accessToken, organizationName, "admin@"+organizationName)

	viewUserRole, err := k.getClientRole(ctx, accessToken, organizationName, realmManagementClientUuid, "view-users")
	if err != nil {
		return realmUUID, err
	}

	err = k.addClientRoleToGroup(ctx, accessToken, organizationName, realmManagementClientUuid, userGroupUuid,
		&gocloak.Role{
			ID:   viewUserRole.ID,
			Name: viewUserRole.Name,
		})

	if err != nil {
		return realmUUID, err
	}

	// TODO: implement leader, member, viewer
	//leaderGroup, err := c.ensureGroup(ctx, token, realmName, "leader@"+realmName)
	//memberGroup, err := c.ensureGroup(ctx, token, realmName, "member@"+realmName)
	//viewerGroup, err := c.ensureGroup(ctx, token, realmName, "viewer@"+realmName)

	return realmUUID, nil
}

func (k *Keycloak) GetRealm(organizationName string, accessToken string) (*domain.Organization, error) {
	ctx := context.Background()
	realm, err := k.client.GetRealm(ctx, accessToken, organizationName)
	if err != nil {
		return nil, err
	}
	return k.reflectOrganization(*realm), nil
}

func (k *Keycloak) GetRealms(accessToken string) ([]*domain.Organization, error) {
	ctx := context.Background()
	realms, err := k.client.GetRealms(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	organization := make([]*domain.Organization, 0)
	for _, realm := range realms {
		organization = append(organization, k.reflectOrganization(*realm))
	}
	return organization, nil
}

func (k *Keycloak) UpdateRealm(organizationName string, organizationConfig domain.Organization, accessToken string) error {
	ctx := context.Background()
	realm := k.reflectRealmRepresentation(organizationConfig)
	err := k.client.UpdateRealm(ctx, accessToken, *realm)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) DeleteRealm(organizationName string, accessToken string) error {
	ctx := context.Background()
	err := k.client.DeleteRealm(ctx, accessToken, organizationName)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) CreateUser(organizationName string, user *gocloak.User, accessToken string) error {
	ctx := context.Background()
	user.Enabled = gocloak.BoolP(true)
	_, err := k.client.CreateUser(ctx, accessToken, organizationName, *user)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) GetUser(organizationName string, accountId string, accessToken string) (*gocloak.User, error) {
	ctx := context.Background()
	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(ctx, accessToken, organizationName, gocloak.GetUsersParams{
		Username: gocloak.StringP(accountId),
	})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("user %s not found", accountId))
	}
	return users[0], nil
}

func (k *Keycloak) GetUsers(organizationName string, accessToken string) ([]*gocloak.User, error) {
	ctx := context.Background()
	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(ctx, accessToken, organizationName, gocloak.GetUsersParams{})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("users not found"))
	}

	return users, nil
}

func (k *Keycloak) UpdateUser(organizationName string, user *gocloak.User, accessToken string) error {
	ctx := context.Background()
	user.Enabled = gocloak.BoolP(true)
	err := k.client.UpdateUser(ctx, accessToken, organizationName, *user)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) DeleteUser(organizationName string, userAccountId string, accessToken string) error {
	ctx := context.Background()
	u, err := k.GetUser(organizationName, userAccountId, accessToken)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		return httpErrors.NewNotFoundError(err)
	}
	err = k.client.DeleteUser(ctx, accessToken, organizationName, *u.ID)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) GetAccessTokenByIdPassword(accountId string, password string, organizationName string) (*domain.User, error) {
	ctx := context.Background()
	JWTToken, err := k.client.Login(ctx, DefaultClientID, DefaultClientSecret, organizationName, accountId, password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &domain.User{
		Token: JWTToken.AccessToken,
	}, nil
}

func (k *Keycloak) VerifyAccessToken(token string, organizationName string) error {

	//TODO implement me
	ctx := context.Background()
	//log.Info(token)
	rptResult, err := k.client.RetrospectToken(ctx, token, DefaultClientID, DefaultClientSecret, organizationName)
	if err != nil {
		return err
	}

	if !(*rptResult.Active) {
		return err
	}
	return nil
}

func (k *Keycloak) ParseAccessToken(token string, organization string) (*jwt.Token, *jwt.MapClaims, error) {
	ctx := context.Background()
	jwtToken, mapClaim, err := k.client.DecodeAccessToken(ctx, token, organization)
	if err != nil {
		fmt.Println(err)
	}

	return jwtToken, mapClaim, err
}

func (c *Keycloak) loginAdmin(ctx context.Context) (*gocloak.JWT, error) {
	token, err := c.client.LoginAdmin(ctx, c.config.AdminId, c.config.AdminPassword, DefaultMasterRealm)
	if err != nil {
		log.Error("Login to keycloak as Admin is failed", err)
	}

	return token, err
}

func (c *Keycloak) ensureClientProtocolMappers(ctx context.Context, token *gocloak.JWT, realm string, clientId string,
	scope string, mapper gocloak.ProtocolMapperRepresentation) error {
	//TODO: Check current logic(if exist, do nothing) is fine
	clients, err := c.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientId,
	})
	if err != nil {
		log.Error("Getting Client is failed", err)
		return err
	}
	if clients[0].ProtocolMappers != nil {
		for _, protocolMapper := range *clients[0].ProtocolMappers {
			if *protocolMapper.Name == *mapper.Name {
				log.Warn("Protocol Mapper already exists", *protocolMapper.Name)
				return nil
			}
		}
	}

	if _, err := c.client.CreateClientProtocolMapper(ctx, token.AccessToken, realm, *clients[0].ID, mapper); err != nil {
		log.Error("Creating Client Protocol Mapper is failed", err)
		return err
	}
	return nil
}

func (c *Keycloak) ensureClient(ctx context.Context, token *gocloak.JWT, realm string, clientId string, secret string) (*gocloak.Client, error) {
	keycloakClient, err := c.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientId,
	})
	if err != nil {
		log.Error("Getting Client is failed", err)
	}

	if len(keycloakClient) == 0 {
		_, err = c.client.CreateClient(ctx, token.AccessToken, realm, gocloak.Client{
			ClientID:                  gocloak.StringP(clientId),
			Enabled:                   gocloak.BoolP(true),
			DirectAccessGrantsEnabled: gocloak.BoolP(true),
		})
		if err != nil {
			log.Error("Creating Client is failed", err)
		}
		keycloakClient, err = c.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
			ClientID: &clientId,
		})
		if err != nil {
			log.Error("Getting Client is failed", err)
		}
	}
	if *keycloakClient[0].Secret != secret {
		log.Warn("Client secret is not matched. Overwrite it")
		keycloakClient[0].Secret = gocloak.StringP(secret)
		if err := c.client.UpdateClient(ctx, token.AccessToken, realm, *keycloakClient[0]); err != nil {
			log.Error("Updating Client is failed", err)
		}
	}

	return keycloakClient[0], err
}

func (c *Keycloak) addUserToGroup(ctx context.Context, token *gocloak.JWT, realm string, userID string, groupID string) error {
	groups, err := c.client.GetUserGroups(ctx, token.AccessToken, realm, userID, gocloak.GetGroupsParams{})
	if err != nil {
		log.Error("Getting User Groups is failed")
	}
	for _, group := range groups {
		if *group.ID == groupID {
			return nil
		}
	}

	err = c.client.AddUserToGroup(ctx, token.AccessToken, realm, userID, groupID)
	if err != nil {
		log.Error("Assigning User to Group is failed", err)
	}
	return err
}

func (c *Keycloak) ensureUserByName(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	user, err := c.ensureUser(ctx, token, realm, userName, password)
	return user, err
}

func (c *Keycloak) ensureUser(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	searchParam := gocloak.GetUsersParams{
		Search: gocloak.StringP(userName),
	}
	users, err := c.client.GetUsers(ctx, token.AccessToken, realm, searchParam)
	if err != nil {
		log.Error("Getting User is failed", err)
	}
	if len(users) == 0 {
		user := gocloak.User{
			Username: gocloak.StringP(userName),
			Enabled:  gocloak.BoolP(true),
			Credentials: &[]gocloak.CredentialRepresentation{
				{
					Type:      gocloak.StringP("password"),
					Value:     gocloak.StringP(password),
					Temporary: gocloak.BoolP(false),
				},
			},
		}
		_, err = c.client.CreateUser(ctx, token.AccessToken, realm, user)
		if err != nil {
			log.Error("Creating User is failed", err)
		}

		users, err = c.client.GetUsers(ctx, token.AccessToken, realm, searchParam)
		if err != nil {
			log.Error("Getting User is failed", err)
		}
	}

	return users[0], err
}

func (c *Keycloak) ensureGroupByName(ctx context.Context, token *gocloak.JWT, realm string, groupName string, groupParam ...gocloak.Group) (*gocloak.Group, error) {
	group, err := c.ensureGroup(ctx, token, realm, groupName)
	return group, err
}

func (c *Keycloak) ensureGroup(ctx context.Context, token *gocloak.JWT, realm string, groupName string) (*gocloak.Group, error) {
	searchParam := gocloak.GetGroupsParams{
		Search: gocloak.StringP(groupName),
	}
	groupParam := gocloak.Group{
		Name: gocloak.StringP(groupName),
	}

	groups, err := c.client.GetGroups(ctx, token.AccessToken, realm, searchParam)
	if err != nil {
		log.Error("Getting Group is failed", err)
	}
	if len(groups) == 0 {
		_, err = c.client.CreateGroup(ctx, token.AccessToken, realm, groupParam)
		if err != nil {
			log.Error("Creating Group is failed", err)
		}
		groups, err = c.client.GetGroups(ctx, token.AccessToken, realm, searchParam)
		if err != nil {
			log.Error("Getting Group is failed", err)
		}
	}

	return groups[0], err
}
func (k *Keycloak) createGroup(ctx context.Context, accessToken string, realm string, groupName string) (string, error) {
	id, err := k.client.CreateGroup(ctx, accessToken, realm, gocloak.Group{Name: gocloak.StringP(groupName)})
	if err != nil {
		log.Error("Creating Group is failed", err)
		return "", err
	}
	return id, nil
}

func (k *Keycloak) getClientByClientId(ctx context.Context, accessToken string, realm string, clientId string) (
	string, error) {
	clients, err := k.client.GetClients(ctx, accessToken, realm, gocloak.GetClientsParams{ClientID: &clientId})
	if err != nil {
		log.Error("Getting Client is failed", err)
		return "", err
	}
	return *clients[0].ID, nil
}

func (k *Keycloak) getClientRole(ctx context.Context, accessToken string, realm string, clientUuid string,
	roleName string) (*gocloak.Role, error) {
	role, err := k.client.GetClientRole(ctx, accessToken, realm, clientUuid, roleName)
	if err != nil {
		log.Error("Getting Client Role is failed", err)
		return nil, err
	}
	return role, nil
}

func (k *Keycloak) addClientRoleToGroup(ctx context.Context, accessToken string, realm string, clientUuid string,
	groupUuid string, role *gocloak.Role) error {
	err := k.client.AddClientRolesToGroup(ctx, accessToken, realm, clientUuid, groupUuid, []gocloak.Role{*role})
	if err != nil {
		log.Error("Adding Client Role to Group is failed", err)
		return err
	}
	return nil
}

func (k *Keycloak) createClientProtocolMapper(ctx context.Context, accessToken string, realm string,
	id string, mapper gocloak.ProtocolMapperRepresentation) (string, error) {
	id, err := k.client.CreateClientProtocolMapper(ctx, accessToken, realm, id, mapper)
	if err != nil {
		log.Error("Creating Client Protocol Mapper is failed", err)
		return "", err
	}

	return id, nil
}

func (k *Keycloak) createDefaultClient(ctx context.Context, accessToken string, realm string, clientId string,
	clientSecret string) (string, error) {
	id, err := k.client.CreateClient(ctx, accessToken, realm, gocloak.Client{
		ClientID:                  gocloak.StringP(clientId),
		DirectAccessGrantsEnabled: gocloak.BoolP(true),
		Enabled:                   gocloak.BoolP(true),
	})

	if err != nil {
		log.Error("Creating Client is failed", err)
		return "", err
	}
	client, err := k.client.GetClient(ctx, accessToken, realm, id)
	if err != nil {
		log.Error("Getting Client is failed", err)
		return "", err
	}
	client.Secret = gocloak.StringP(clientSecret)
	err = k.client.UpdateClient(ctx, accessToken, realm, *client)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (k *Keycloak) reflectOrganization(org gocloak.RealmRepresentation) *domain.Organization {
	return &domain.Organization{
		ID:   *org.ID,
		Name: *org.Realm,
	}
}

func (k *Keycloak) reflectRealmRepresentation(org domain.Organization) *gocloak.RealmRepresentation {
	return &gocloak.RealmRepresentation{
		Realm: gocloak.StringP(org.Name),
	}
}

// var defaultProtocolTksMapper = make([]gocloak.ProtocolMapperRepresentation, 3)
var defaultProtocolTksMapper = []gocloak.ProtocolMapperRepresentation{
	{
		Name:            gocloak.StringP("org"),
		Protocol:        gocloak.StringP("openid-connect"),
		ProtocolMapper:  gocloak.StringP("oidc-hardcoded-claim-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config: &map[string]string{
			"full.path":            "false",
			"id.token.claim":       "false",
			"access.token.claim":   "true",
			"claim.name":           "organization",
			"claim.value":          "master",
			"userinfo.token.claim": "false",
		},
	},
	{
		Name:            gocloak.StringP("tksrole"),
		Protocol:        gocloak.StringP("openid-connect"),
		ProtocolMapper:  gocloak.StringP("oidc-group-membership-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config: &map[string]string{
			"full.path":            "false",
			"id.token.claim":       "false",
			"access.token.claim":   "true",
			"claim.name":           "tks-role",
			"userinfo.token.claim": "false",
		},
	},
	{
		Name:            gocloak.StringP("aud"),
		Protocol:        gocloak.StringP("openid-connect"),
		ProtocolMapper:  gocloak.StringP("oidc-claims-param-token-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config: &map[string]string{
			"id.token.claim":       "true",
			"userinfo.token.claim": "false",
		},
	},
}
