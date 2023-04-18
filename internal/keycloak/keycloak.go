package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/openinfradev/tks-api/pkg/httpErrors"

	"github.com/Nerzal/gocloak/v13"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IKeycloak interface {
	InitializeKeycloak() error

	LoginAdmin(accountId string, password string) (*domain.User, error)
	Login(accountId string, password string, organizationId string) (*domain.User, error)

	CreateRealm(organizationName string) (string, error)
	GetRealm(organizationName string) (*domain.Organization, error)
	GetRealms() ([]*domain.Organization, error)
	DeleteRealm(organizationName string) error
	UpdateRealm(organizationName string, organizationConfig domain.Organization) error

	CreateUser(organizationName string, user *gocloak.User) error
	GetUser(organizationName string, userAccountId string) (*gocloak.User, error)
	GetUsers(organizationName string) ([]*gocloak.User, error)
	DeleteUser(organizationName string, userAccountId string) error
	UpdateUser(organizationName string, user *gocloak.User) error

	VerifyAccessToken(token string, organizationName string) error
}
type Keycloak struct {
	config *Config
	client *gocloak.GoCloak
}

func (k *Keycloak) LoginAdmin(accountId string, password string) (*domain.User, error) {
	ctx := context.Background()
	log.Info("LoginAdmin called")
	//JWTToken, err := k.client.Login(ctx, "admin-cli", "", DefaultMasterRealm, accountId, password)
	JWTToken, err := k.client.LoginAdmin(ctx, accountId, password, DefaultMasterRealm)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &domain.User{Token: JWTToken.AccessToken}, nil
}

func (k *Keycloak) Login(accountId string, password string, organizationId string) (*domain.User, error) {
	ctx := context.Background()
	JWTToken, err := k.client.Login(ctx, DefaultClientID, DefaultClientSecret, organizationId, accountId, password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &domain.User{Token: JWTToken.AccessToken}, nil
}

func New(config *Config) IKeycloak {
	return &Keycloak{
		config: config,
	}
}
func (k *Keycloak) InitializeKeycloak() error {
	k.client = gocloak.NewClient(k.config.Address)
	ctx := context.Background()
	restyClient := k.client.RestyClient()
	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	token, err := k.loginAdmin(ctx)
	if err != nil {
		log.Fatal(err)
		return err
	}

	group, err := k.ensureGroupByName(ctx, token, DefaultMasterRealm, "tks-admin@master")
	if err != nil {
		log.Fatal(err)
		return err
	}

	user, err := k.ensureUserByName(ctx, token, DefaultMasterRealm, k.config.AdminId, k.config.AdminPassword)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err := k.addUserToGroup(ctx, token, DefaultMasterRealm, *user.ID, *group.ID); err != nil {
		log.Fatal(err)
		return err
	}

	keycloakClient, err := k.ensureClient(ctx, token, DefaultMasterRealm, DefaultClientID, DefaultClientSecret)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, defaultMapper := range defaultProtocolTksMapper {
		if err := k.ensureClientProtocolMappers(ctx, token, DefaultMasterRealm, *keycloakClient.ClientID, "openid", defaultMapper); err != nil {
			log.Fatal(err)
			return err
		}
	}

	if _, err := k.client.Login(ctx, DefaultClientID, DefaultClientSecret, DefaultMasterRealm,
		k.config.AdminId, k.config.AdminPassword); err != nil {
		log.Fatal(err)
		return err
	}

	adminCliClient, err := k.ensureClient(ctx, token, DefaultMasterRealm, AdminCliClientID, DefaultClientSecret)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, defaultMapper := range defaultProtocolTksMapper {
		if err := k.ensureClientProtocolMappers(ctx, token, DefaultMasterRealm, *adminCliClient.ClientID, "openid", defaultMapper); err != nil {
			log.Fatal(err)
			return err
		}
	}

	if _, err := k.client.Login(ctx, AdminCliClientID, DefaultClientSecret, DefaultMasterRealm,
		k.config.AdminId, k.config.AdminPassword); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (k *Keycloak) CreateRealm(organizationName string) (string, error) {
	//TODO implement me
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return "", err
	}
	accessToken := token.AccessToken

	realmConfig := gocloak.RealmRepresentation{
		Realm:               &organizationName,
		Enabled:             gocloak.BoolP(true),
		AccessTokenLifespan: gocloak.IntP(accessTokenLifespan),
	}
	realmUUID, err := k.client.CreateRealm(ctx, accessToken, realmConfig)
	if err != nil {
		return "", err
	}

	// After Create Realm, accesstoken got changed so that old token doesn't work properly.
	token, err = k.loginAdmin(ctx)
	if err != nil {
		return "", err
	}
	accessToken = token.AccessToken

	clientUUID, err := k.createDefaultClient(context.Background(), accessToken, organizationName, DefaultClientID, DefaultClientSecret)
	if err != nil {
		log.Error(err, "createDefaultClient")
		return "", err
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
			return "", err
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
		return "", err
	}

	userGroupUuid, err := k.createGroup(ctx, accessToken, organizationName, "user@"+organizationName)
	if err != nil {
		return "", err
	}
	//adminGroup, err := k.ensureGroup(ctx, accessToken, organizationName, "admin@"+organizationName)

	viewUserRole, err := k.getClientRole(ctx, accessToken, organizationName, realmManagementClientUuid, "view-users")
	if err != nil {
		return "", err
	}

	err = k.addClientRoleToGroup(ctx, accessToken, organizationName, realmManagementClientUuid, userGroupUuid,
		&gocloak.Role{
			ID:   viewUserRole.ID,
			Name: viewUserRole.Name,
		})

	if err != nil {
		return "", err
	}

	// TODO: implement leader, member, viewer
	//leaderGroup, err := c.ensureGroup(ctx, token, realmName, "leader@"+realmName)
	//memberGroup, err := c.ensureGroup(ctx, token, realmName, "member@"+realmName)
	//viewerGroup, err := c.ensureGroup(ctx, token, realmName, "viewer@"+realmName)

	return realmUUID, nil
}

func (k *Keycloak) GetRealm(organizationName string) (*domain.Organization, error) {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return nil, err
	}
	realm, err := k.client.GetRealm(ctx, token.AccessToken, organizationName)
	if err != nil {
		return nil, err
	}
	return k.reflectOrganization(*realm), nil
}

func (k *Keycloak) GetRealms() ([]*domain.Organization, error) {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return nil, err
	}
	realms, err := k.client.GetRealms(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	organization := make([]*domain.Organization, 0)
	for _, realm := range realms {
		organization = append(organization, k.reflectOrganization(*realm))
	}
	return organization, nil
}

func (k *Keycloak) UpdateRealm(organizationName string, organizationConfig domain.Organization) error {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return err
	}
	realm := k.reflectRealmRepresentation(organizationConfig)
	err = k.client.UpdateRealm(ctx, token.AccessToken, *realm)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) DeleteRealm(organizationName string) error {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return err
	}
	err = k.client.DeleteRealm(ctx, token.AccessToken, organizationName)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) CreateUser(organizationName string, user *gocloak.User) error {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return err
	}
	user.Enabled = gocloak.BoolP(true)
	_, err = k.client.CreateUser(ctx, token.AccessToken, organizationName, *user)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) GetUser(organizationName string, accountId string) (*gocloak.User, error) {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return nil, err
	}

	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(ctx, token.AccessToken, organizationName, gocloak.GetUsersParams{
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

func (k *Keycloak) GetUsers(organizationName string) ([]*gocloak.User, error) {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return nil, err
	}
	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(ctx, token.AccessToken, organizationName, gocloak.GetUsersParams{})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("users not found"))
	}

	return users, nil
}

func (k *Keycloak) UpdateUser(organizationName string, user *gocloak.User) error {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return err
	}
	user.Enabled = gocloak.BoolP(true)
	err = k.client.UpdateUser(ctx, token.AccessToken, organizationName, *user)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) DeleteUser(organizationName string, userAccountId string) error {
	ctx := context.Background()
	token, err := k.loginAdmin(ctx)
	if err != nil {
		return err
	}
	u, err := k.GetUser(organizationName, userAccountId)
	if err != nil {
		log.Errorf("error is :%s(%T)", err.Error(), err)
		return httpErrors.NewNotFoundError(err)
	}
	err = k.client.DeleteUser(ctx, token.AccessToken, organizationName, *u.ID)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) VerifyAccessToken(token string, organizationName string) error {
	ctx := context.Background()
	rptResult, err := k.client.RetrospectToken(ctx, token, DefaultClientID, DefaultClientSecret, organizationName)
	if err != nil {
		return err
	}

	if !(*rptResult.Active) {
		return err
	}
	return nil
}

func (k *Keycloak) loginAdmin(ctx context.Context) (*gocloak.JWT, error) {
	token, err := k.client.LoginAdmin(ctx, k.config.AdminId, k.config.AdminPassword, DefaultMasterRealm)
	if err != nil {
		log.Error("Login to keycloak as Admin is failed", err)
	}

	return token, err
}

func (k *Keycloak) ensureClientProtocolMappers(ctx context.Context, token *gocloak.JWT, realm string, clientId string,
	scope string, mapper gocloak.ProtocolMapperRepresentation) error {
	//TODO: Check current logic(if exist, do nothing) is fine
	clients, err := k.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
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

	if _, err := k.client.CreateClientProtocolMapper(ctx, token.AccessToken, realm, *clients[0].ID, mapper); err != nil {
		log.Error("Creating Client Protocol Mapper is failed", err)
		return err
	}
	return nil
}

func (k *Keycloak) ensureClient(ctx context.Context, token *gocloak.JWT, realm string, clientId string, secret string) (*gocloak.Client, error) {
	keycloakClient, err := k.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientId,
	})
	if err != nil {
		log.Error("Getting Client is failed", err)
	}

	if len(keycloakClient) == 0 {
		_, err = k.client.CreateClient(ctx, token.AccessToken, realm, gocloak.Client{
			ClientID:                  gocloak.StringP(clientId),
			Enabled:                   gocloak.BoolP(true),
			DirectAccessGrantsEnabled: gocloak.BoolP(true),
		})
		if err != nil {
			log.Error("Creating Client is failed", err)
		}
		keycloakClient, err = k.client.GetClients(ctx, token.AccessToken, realm, gocloak.GetClientsParams{
			ClientID: &clientId,
		})
		if err != nil {
			log.Error("Getting Client is failed", err)
		}
	}
	if keycloakClient[0].Secret == nil || *keycloakClient[0].Secret != secret {
		log.Warn("Client secret is not matched. Overwrite it")
		keycloakClient[0].Secret = gocloak.StringP(secret)
		if err := k.client.UpdateClient(ctx, token.AccessToken, realm, *keycloakClient[0]); err != nil {
			log.Error("Updating Client is failed", err)
		}
	}

	return keycloakClient[0], nil
}

func (k *Keycloak) addUserToGroup(ctx context.Context, token *gocloak.JWT, realm string, userID string, groupID string) error {
	groups, err := k.client.GetUserGroups(ctx, token.AccessToken, realm, userID, gocloak.GetGroupsParams{})
	if err != nil {
		log.Error("Getting User Groups is failed")
	}
	for _, group := range groups {
		if *group.ID == groupID {
			return nil
		}
	}

	err = k.client.AddUserToGroup(ctx, token.AccessToken, realm, userID, groupID)
	if err != nil {
		log.Error("Assigning User to Group is failed", err)
	}
	return err
}

func (k *Keycloak) ensureUserByName(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	user, err := k.ensureUser(ctx, token, realm, userName, password)
	return user, err
}

func (k *Keycloak) ensureUser(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	searchParam := gocloak.GetUsersParams{
		Search: gocloak.StringP(userName),
	}
	users, err := k.client.GetUsers(ctx, token.AccessToken, realm, searchParam)
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
		_, err = k.client.CreateUser(ctx, token.AccessToken, realm, user)
		if err != nil {
			log.Error("Creating User is failed", err)
		}

		users, err = k.client.GetUsers(ctx, token.AccessToken, realm, searchParam)
		if err != nil {
			log.Error("Getting User is failed", err)
		}
	}

	return users[0], err
}

func (k *Keycloak) ensureGroupByName(ctx context.Context, token *gocloak.JWT, realm string, groupName string, groupParam ...gocloak.Group) (*gocloak.Group, error) {
	group, err := k.ensureGroup(ctx, token, realm, groupName)
	return group, err
}

func (k *Keycloak) ensureGroup(ctx context.Context, token *gocloak.JWT, realm string, groupName string) (*gocloak.Group, error) {
	searchParam := gocloak.GetGroupsParams{
		Search: gocloak.StringP(groupName),
	}
	groupParam := gocloak.Group{
		Name: gocloak.StringP(groupName),
	}

	groups, err := k.client.GetGroups(ctx, token.AccessToken, realm, searchParam)
	if err != nil {
		log.Error("Getting Group is failed", err)
	}
	if len(groups) == 0 {
		_, err = k.client.CreateGroup(ctx, token.AccessToken, realm, groupParam)
		if err != nil {
			log.Error("Creating Group is failed", err)
		}
		groups, err = k.client.GetGroups(ctx, token.AccessToken, realm, searchParam)
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
