package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/spf13/viper"

	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IKeycloak interface {
	InitializeKeycloak(ctx context.Context) error

	LoginAdmin(ctx context.Context, accountId string, password string) (*model.User, error)
	Login(ctx context.Context, accountId string, password string, organizationId string) (*model.User, error)
	Logout(ctx context.Context, sessionId string, organizationId string) error

	CreateRealm(ctx context.Context, organizationId string) (string, error)
	GetRealm(ctx context.Context, organizationId string) (*model.Organization, error)
	GetRealms(ctx context.Context) ([]*model.Organization, error)
	DeleteRealm(ctx context.Context, organizationId string) error
	UpdateRealm(ctx context.Context, organizationId string, organizationConfig model.Organization) error

	CreateUser(ctx context.Context, organizationId string, user *gocloak.User) (string, error)
	GetUser(ctx context.Context, organizationId string, userAccountId string) (*gocloak.User, error)
	GetUsers(ctx context.Context, organizationId string) ([]*gocloak.User, error)
	DeleteUser(ctx context.Context, organizationId string, userAccountId string) error
	UpdateUser(ctx context.Context, organizationId string, user *gocloak.User) error
	JoinGroup(ctx context.Context, organizationId string, userId string, groupName string) error
	LeaveGroup(ctx context.Context, organizationId string, userId string, groupName string) error

	EnsureClientRoleWithClientName(ctx context.Context, organizationId string, clientName string, roleName string) error
	DeleteClientRoleWithClientName(ctx context.Context, organizationId string, clientName string, roleName string) error

	AssignClientRoleToUser(ctx context.Context, organizationId string, userId string, clientName string, roleName string) error
	UnassignClientRoleToUser(ctx context.Context, organizationId string, userId string, clientName string, roleName string) error

	VerifyAccessToken(ctx context.Context, token string, organizationId string) (bool, error)
	GetSessions(ctx context.Context, userId string, organizationId string) (*[]string, error)
}
type Keycloak struct {
	config        *Config
	client        *gocloak.GoCloak
	adminCliToken *gocloak.JWT
}

func (k *Keycloak) LoginAdmin(ctx context.Context, accountId string, password string) (*model.User, error) {
	JWTToken, err := k.client.LoginAdmin(context.Background(), accountId, password, DefaultMasterRealm)
	if err != nil {
		log.Error(ctx, err)
		return nil, err
	}

	return &model.User{Token: JWTToken.AccessToken}, nil
}

func (k *Keycloak) Login(ctx context.Context, accountId string, password string, organizationId string) (*model.User, error) {
	JWTToken, err := k.client.Login(context.Background(), DefaultClientID, k.config.ClientSecret, organizationId, accountId, password)
	if err != nil {
		log.Error(ctx, err)
		return nil, err
	}
	return &model.User{Token: JWTToken.AccessToken}, nil
}

func New(config *Config) IKeycloak {
	return &Keycloak{
		config: config,
	}
}
func (k *Keycloak) InitializeKeycloak(ctx context.Context) error {
	k.client = gocloak.NewClient(k.config.Address)
	restyClient := k.client.RestyClient()
	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	var token *gocloak.JWT
	var err error
	if token, err = k.client.LoginAdmin(context.Background(), k.config.AdminId, k.config.AdminPassword, DefaultMasterRealm); err != nil {
		log.Fatal(ctx, err)
		return err
	}
	k.adminCliToken = token

	err = k.client.UpdateRealm(ctx, token.AccessToken, defaultRealmSetting(DefaultMasterRealm))
	if err != nil {
		return err
	}

	group, err := k.ensureGroupByName(ctx, token, DefaultMasterRealm, "tks-admin@master")
	if err != nil {
		log.Fatal(ctx, err)
		return err
	}

	user, err := k.ensureUserByName(ctx, token, DefaultMasterRealm, k.config.AdminId, k.config.AdminPassword)
	if err != nil {
		log.Fatal(ctx, err)
		return err
	}

	if err := k.addUserToGroup(ctx, token, DefaultMasterRealm, *user.ID, *group.ID); err != nil {
		log.Fatal(ctx, err)
		return err
	}

	var redirectURIs []string
	redirectURIs = append(redirectURIs, viper.GetString("external-address")+"/*")
	tksClient, err := k.ensureClient(ctx, token, DefaultMasterRealm, DefaultClientID, k.config.ClientSecret, &redirectURIs)
	if err != nil {
		log.Fatal(ctx, err)
		return err
	}

	//
	for _, defaultMapper := range defaultProtocolTksMapper {
		if err := k.ensureClientProtocolMappers(ctx, token, DefaultMasterRealm, *tksClient.ClientID, "openid", defaultMapper); err != nil {
			log.Fatal(ctx, err)
			return err
		}
	}

	adminCliClient, err := k.ensureClient(ctx, token, DefaultMasterRealm, AdminCliClientID, k.config.ClientSecret, nil)
	if err != nil {
		log.Fatal(ctx, err)
		return err
	}

	for _, defaultMapper := range defaultProtocolTksMapper {
		if err := k.ensureClientProtocolMappers(ctx, token, DefaultMasterRealm, *adminCliClient.ClientID, "openid", defaultMapper); err != nil {
			log.Fatal(ctx, err)
			return err
		}
	}

	// Todo: 현재 30초마다 갱신하도록 함. 최적화 요소 확인 및 개선 필요
	_ = getRefreshTokenExpiredDuration(k.adminCliToken)
	go func() {
		for {
			if token, err := k.client.RefreshToken(context.Background(), k.adminCliToken.RefreshToken, AdminCliClientID, k.config.ClientSecret, DefaultMasterRealm); err != nil {
				log.Errorf(ctx, "[Refresh]error is :%s(%T)", err.Error(), err)
				log.Info(ctx, "[Do Keycloak Admin CLI Login]")
				k.adminCliToken, err = k.client.LoginAdmin(ctx, k.config.AdminId, k.config.AdminPassword, DefaultMasterRealm)
				if err != nil {
					log.Errorf(ctx, "[LoginAdmin]error is :%s(%T)", err.Error(), err)
				}
			} else {
				k.adminCliToken = token
			}
			time.Sleep(30 * time.Second)
		}
	}()

	return nil
}

func (k *Keycloak) CreateRealm(ctx context.Context, organizationId string) (string, error) {
	//TODO implement me
	token := k.adminCliToken

	realmUUID, err := k.client.CreateRealm(context.Background(), token.AccessToken, defaultRealmSetting(organizationId))
	if err != nil {
		return "", err
	}

	var redirectURIs []string
	redirectURIs = append(redirectURIs, viper.GetString("external-address")+"/*")
	clientUUID, err := k.createDefaultClient(context.Background(), token.AccessToken, organizationId, DefaultClientID, k.config.ClientSecret, &redirectURIs)
	if err != nil {
		log.Error(ctx, err, "createDefaultClient")
		return "", err
	}

	for _, defaultMapper := range defaultProtocolTksMapper {
		if *defaultMapper.Name == "org" {
			defaultMapper.Config = &map[string]string{
				"full.path":            "false",
				"id.token.claim":       "false",
				"access.token.claim":   "true",
				"claim.name":           "organization",
				"claim.value":          organizationId,
				"userinfo.token.claim": "false",
			}
		}
		if _, err := k.createClientProtocolMapper(context.Background(), token.AccessToken, organizationId, clientUUID, defaultMapper); err != nil {
			return "", err
		}
	}
	adminGroupUuid, err := k.createGroup(ctx, token.AccessToken, organizationId, "admin@"+organizationId)
	if err != nil {
		return realmUUID, err
	}

	realmManagementClientUuid, err := k.getClientByClientId(ctx, token.AccessToken, organizationId, "realm-management")
	if err != nil {
		return realmUUID, err
	}

	realmAdminRole, err := k.getClientRole(ctx, token.AccessToken, organizationId, realmManagementClientUuid, "realm-admin")
	if err != nil {
		return realmUUID, err
	}

	err = k.addClientRoleToGroup(ctx, token.AccessToken, organizationId, realmManagementClientUuid, adminGroupUuid,
		&gocloak.Role{
			ID:   realmAdminRole.ID,
			Name: realmAdminRole.Name,
		})

	if err != nil {
		return "", err
	}

	userGroupUuid, err := k.createGroup(ctx, token.AccessToken, organizationId, "user@"+organizationId)
	if err != nil {
		return "", err
	}

	viewUserRole, err := k.getClientRole(ctx, token.AccessToken, organizationId, realmManagementClientUuid, "view-users")
	if err != nil {
		return "", err
	}

	err = k.addClientRoleToGroup(ctx, token.AccessToken, organizationId, realmManagementClientUuid, userGroupUuid,
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

func (k *Keycloak) GetRealm(ctx context.Context, organizationId string) (*model.Organization, error) {
	token := k.adminCliToken
	realm, err := k.client.GetRealm(context.Background(), token.AccessToken, organizationId)
	if err != nil {
		return nil, err
	}

	return k.reflectOrganization(*realm), nil
}

func (k *Keycloak) GetRealms(ctx context.Context) ([]*model.Organization, error) {
	token := k.adminCliToken
	realms, err := k.client.GetRealms(context.Background(), token.AccessToken)
	if err != nil {
		return nil, err
	}
	organization := make([]*model.Organization, 0)
	for _, realm := range realms {
		organization = append(organization, k.reflectOrganization(*realm))
	}

	return organization, nil
}

func (k *Keycloak) UpdateRealm(ctx context.Context, organizationId string, organizationConfig model.Organization) error {
	token := k.adminCliToken
	realm := k.reflectRealmRepresentation(organizationConfig)
	err := k.client.UpdateRealm(context.Background(), token.AccessToken, *realm)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) DeleteRealm(ctx context.Context, organizationId string) error {
	token := k.adminCliToken
	err := k.client.DeleteRealm(context.Background(), token.AccessToken, organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keycloak) CreateUser(ctx context.Context, organizationId string, user *gocloak.User) (string, error) {
	token := k.adminCliToken
	user.Enabled = gocloak.BoolP(true)
	uuid, err := k.client.CreateUser(context.Background(), token.AccessToken, organizationId, *user)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (k *Keycloak) GetUser(ctx context.Context, organizationId string, accountId string) (*gocloak.User, error) {
	token := k.adminCliToken

	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(context.Background(), token.AccessToken, organizationId, gocloak.GetUsersParams{
		Username: gocloak.StringP(accountId),
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("user %s not found", accountId), "", "")
	}

	return users[0], nil
}

func (k *Keycloak) GetUsers(ctx context.Context, organizationId string) ([]*gocloak.User, error) {
	token := k.adminCliToken
	//TODO: this is rely on the fact that username is the same as userAccountId and unique
	users, err := k.client.GetUsers(context.Background(), token.AccessToken, organizationId, gocloak.GetUsersParams{})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf("users not found"), "", "")
	}

	return users, nil
}

func (k *Keycloak) UpdateUser(ctx context.Context, organizationId string, user *gocloak.User) error {
	token := k.adminCliToken
	user.Enabled = gocloak.BoolP(true)
	err := k.client.UpdateUser(context.Background(), token.AccessToken, organizationId, *user)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keycloak) DeleteUser(ctx context.Context, organizationId string, userAccountId string) error {
	token := k.adminCliToken
	u, err := k.GetUser(ctx, organizationId, userAccountId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		return httpErrors.NewNotFoundError(err, "", "")
	}
	err = k.client.DeleteUser(context.Background(), token.AccessToken, organizationId, *u.ID)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keycloak) VerifyAccessToken(ctx context.Context, token string, organizationId string) (bool, error) {
	rptResult, err := k.client.RetrospectToken(context.Background(), token, DefaultClientID, k.config.ClientSecret, organizationId)
	if err != nil {
		return false, err
	}
	if !(*rptResult.Active) {
		return false, nil
	}

	return true, nil
}

func (k *Keycloak) GetSessions(ctx context.Context, userId string, organizationId string) (*[]string, error) {
	token := k.adminCliToken
	sessions, err := k.client.GetUserSessions(context.Background(), token.AccessToken, organizationId, userId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		return nil, err
	}

	var sessionIds []string
	for _, session := range sessions {
		sessionIds = append(sessionIds, *session.ID)
	}

	return &sessionIds, nil
}

func (k *Keycloak) Logout(ctx context.Context, sessionId string, organizationId string) error {
	token := k.adminCliToken
	err := k.client.LogoutUserSession(context.Background(), token.AccessToken, organizationId, sessionId)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keycloak) JoinGroup(ctx context.Context, organizationId string, userId string, groupName string) error {
	token := k.adminCliToken
	groups, err := k.client.GetGroups(context.Background(), token.AccessToken, organizationId, gocloak.GetGroupsParams{
		Search: &groupName,
	})
	if err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "", "")
	}
	if len(groups) == 0 {
		return httpErrors.NewNotFoundError(fmt.Errorf("group not found"), "", "")
	}
	if err := k.client.AddUserToGroup(ctx, token.AccessToken, organizationId, userId, *groups[0].ID); err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (k *Keycloak) LeaveGroup(ctx context.Context, organizationId string, userId string, groupName string) error {
	token := k.adminCliToken
	groups, err := k.client.GetGroups(context.Background(), token.AccessToken, organizationId, gocloak.GetGroupsParams{
		Search: &groupName,
	})
	if err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "", "")
	}
	if len(groups) == 0 {
		return httpErrors.NewNotFoundError(fmt.Errorf("group not found"), "", "")
	}
	if err := k.client.DeleteUserFromGroup(context.Background(), token.AccessToken, organizationId, userId, *groups[0].ID); err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (k *Keycloak) EnsureClientRoleWithClientName(ctx context.Context, organizationId string, clientName string, roleName string) error {
	token := k.adminCliToken

	clients, err := k.client.GetClients(context.Background(), token.AccessToken, organizationId, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return err
	}

	targetClient := clients[0]

	role := gocloak.Role{
		Name: gocloak.StringP(roleName),
	}

	_, err = k.client.CreateClientRole(context.Background(), token.AccessToken, organizationId, *targetClient.ID, role)
	if err != nil {
		log.Error(ctx, "Creating Client Role is failed", err)
		return err
	}

	return nil
}

func (k *Keycloak) DeleteClientRoleWithClientName(ctx context.Context, organizationId string, clientName string, roleName string) error {
	token := k.adminCliToken

	clients, err := k.client.GetClients(context.Background(), token.AccessToken, organizationId, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return err
	}

	targetClient := clients[0]

	roles, err := k.client.GetClientRoles(context.Background(), token.AccessToken, organizationId, *targetClient.ID, gocloak.GetRoleParams{
		Search: &roleName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client Role is failed", err)
		return err
	}

	if len(roles) == 0 {
		log.Warn(ctx, "Client Role not found", roleName)
		return nil
	}

	err = k.client.DeleteClientRole(context.Background(), token.AccessToken, organizationId, *targetClient.ID, *roles[0].ID)
	if err != nil {
		log.Error(ctx, "Deleting Client Role is failed", err)
		return err
	}

	return nil
}

func (k *Keycloak) AssignClientRoleToUser(ctx context.Context, organizationId string, userId string, clientName string, roleName string) error {
	token := k.adminCliToken

	clients, err := k.client.GetClients(context.Background(), token.AccessToken, organizationId, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return err
	}
	if len(clients) == 0 {
		log.Warn(ctx, "Client not found", clientName)
		return nil
	}

	targetClient := clients[0]

	roles, err := k.client.GetClientRoles(context.Background(), token.AccessToken, organizationId, *targetClient.ID, gocloak.GetRoleParams{
		Search: &roleName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client Role is failed", err)
		return err
	}

	if len(roles) == 0 {
		log.Warn(ctx, "Client Role not found", roleName)
		return nil
	}

	err = k.client.AddClientRolesToUser(context.Background(), token.AccessToken, organizationId, userId, *targetClient.ID, []gocloak.Role{*roles[0]})

	if err != nil {
		log.Error(ctx, "Assigning Client Role to User is failed", err)
		return err
	}

	return nil
}

func (k *Keycloak) UnassignClientRoleToUser(ctx context.Context, organizationId string, userId string, clientName string, roleName string) error {
	token := k.adminCliToken

	clients, err := k.client.GetClients(context.Background(), token.AccessToken, organizationId, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return err
	}
	if len(clients) == 0 {
		log.Warn(ctx, "Client not found", clientName)
		return nil
	}

	targetClient := clients[0]

	roles, err := k.client.GetClientRoles(context.Background(), token.AccessToken, organizationId, *targetClient.ID, gocloak.GetRoleParams{
		Search: &roleName,
	})
	if err != nil {
		log.Error(ctx, "Getting Client Role is failed", err)
		return err
	}

	if len(roles) == 0 {
		log.Warn(ctx, "Client Role not found", roleName)
		return nil
	}

	err = k.client.DeleteClientRolesFromUser(context.Background(), token.AccessToken, organizationId, userId, *targetClient.ID, []gocloak.Role{*roles[0]})
	if err != nil {
		log.Error(ctx, "Unassigning Client Role to User is failed", err)
		return err
	}

	return nil
}

func (k *Keycloak) ensureClientProtocolMappers(ctx context.Context, token *gocloak.JWT, realm string, clientId string,
	scope string, mapper gocloak.ProtocolMapperRepresentation) error {
	//TODO: Check current logic(if exist, do nothing) is fine
	clients, err := k.client.GetClients(context.Background(), token.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientId,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return err
	}
	if clients[0].ProtocolMappers != nil {
		for _, protocolMapper := range *clients[0].ProtocolMappers {
			if *protocolMapper.Name == *mapper.Name {
				log.Warn(ctx, "Protocol Mapper already exists", *protocolMapper.Name)
				return nil
			}
		}
	}

	if _, err := k.client.CreateClientProtocolMapper(context.Background(), token.AccessToken, realm, *clients[0].ID, mapper); err != nil {
		log.Error(ctx, "Creating Client Protocol Mapper is failed", err)
		return err
	}
	return nil
}

func (k *Keycloak) ensureClient(ctx context.Context, token *gocloak.JWT, realm string, clientId string, secret string, redirectURIs *[]string) (*gocloak.Client, error) {
	keycloakClient, err := k.client.GetClients(context.Background(), token.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientId,
	})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
	}

	if len(keycloakClient) == 0 {
		_, err = k.client.CreateClient(context.Background(), token.AccessToken, realm, gocloak.Client{
			ClientID:                  gocloak.StringP(clientId),
			Enabled:                   gocloak.BoolP(true),
			DirectAccessGrantsEnabled: gocloak.BoolP(true),
			RedirectURIs:              redirectURIs,
		})
		if err != nil {
			log.Error(ctx, "Creating Client is failed", err)
		}
		keycloakClient, err = k.client.GetClients(context.Background(), token.AccessToken, realm, gocloak.GetClientsParams{
			ClientID: &clientId,
		})
		if err != nil {
			log.Error(ctx, "Getting Client is failed", err)
		}
	} else {
		err = k.client.UpdateClient(context.Background(), token.AccessToken, realm, gocloak.Client{
			ID:                        keycloakClient[0].ID,
			Enabled:                   gocloak.BoolP(true),
			DirectAccessGrantsEnabled: gocloak.BoolP(true),
			RedirectURIs:              redirectURIs,
		})
		if err != nil {
			log.Error(ctx, "Update Client is failed", err)
		}
	}
	if keycloakClient[0].Secret == nil || *keycloakClient[0].Secret != secret {
		log.Warn(ctx, "Client secret is not matched. Overwrite it")
		keycloakClient[0].Secret = gocloak.StringP(secret)
		if err := k.client.UpdateClient(context.Background(), token.AccessToken, realm, *keycloakClient[0]); err != nil {
			log.Error(ctx, "Updating Client is failed", err)
		}
	}

	return keycloakClient[0], nil
}

func (k *Keycloak) addUserToGroup(ctx context.Context, token *gocloak.JWT, realm string, userID string, groupID string) error {
	groups, err := k.client.GetUserGroups(context.Background(), token.AccessToken, realm, userID, gocloak.GetGroupsParams{})
	if err != nil {
		log.Error(ctx, "Getting User Groups is failed")
	}
	for _, group := range groups {
		if *group.ID == groupID {
			return nil
		}
	}

	err = k.client.AddUserToGroup(context.Background(), token.AccessToken, realm, userID, groupID)
	if err != nil {
		log.Error(ctx, "Assigning User to Group is failed", err)
	}
	return err
}

func (k *Keycloak) ensureUserByName(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	user, err := k.ensureUser(context.Background(), token, realm, userName, password)
	return user, err
}

func (k *Keycloak) ensureUser(ctx context.Context, token *gocloak.JWT, realm string, userName string, password string) (*gocloak.User, error) {
	searchParam := gocloak.GetUsersParams{
		Search: gocloak.StringP(userName),
	}
	users, err := k.client.GetUsers(context.Background(), token.AccessToken, realm, searchParam)
	if err != nil {
		log.Error(ctx, "Getting User is failed", err)
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
		_, err = k.client.CreateUser(context.Background(), token.AccessToken, realm, user)
		if err != nil {
			log.Error(ctx, "Creating User is failed", err)
		}

		users, err = k.client.GetUsers(context.Background(), token.AccessToken, realm, searchParam)
		if err != nil {
			log.Error(ctx, "Getting User is failed", err)
		}
	}

	return users[0], err
}

func (k *Keycloak) ensureGroupByName(ctx context.Context, token *gocloak.JWT, realm string, groupName string, groupParam ...gocloak.Group) (*gocloak.Group, error) {
	group, err := k.ensureGroup(context.Background(), token, realm, groupName)
	return group, err
}

func (k *Keycloak) ensureGroup(ctx context.Context, token *gocloak.JWT, realm string, groupName string) (*gocloak.Group, error) {
	searchParam := gocloak.GetGroupsParams{
		Search: gocloak.StringP(groupName),
	}
	groupParam := gocloak.Group{
		Name: gocloak.StringP(groupName),
	}

	groups, err := k.client.GetGroups(context.Background(), token.AccessToken, realm, searchParam)
	if err != nil {
		log.Error(ctx, "Getting Group is failed", err)
	}
	if len(groups) == 0 {
		_, err = k.client.CreateGroup(context.Background(), token.AccessToken, realm, groupParam)
		if err != nil {
			log.Error(ctx, "Creating Group is failed", err)
		}
		groups, err = k.client.GetGroups(context.Background(), token.AccessToken, realm, searchParam)
		if err != nil {
			log.Error(ctx, "Getting Group is failed", err)
		}
	}

	return groups[0], err
}
func (k *Keycloak) createGroup(ctx context.Context, accessToken string, realm string, groupName string) (string, error) {
	id, err := k.client.CreateGroup(context.Background(), accessToken, realm, gocloak.Group{Name: gocloak.StringP(groupName)})
	if err != nil {
		log.Error(ctx, "Creating Group is failed", err)
		return "", err
	}
	return id, nil
}

func (k *Keycloak) getClientByClientId(ctx context.Context, accessToken string, realm string, clientId string) (
	string, error) {
	clients, err := k.client.GetClients(context.Background(), accessToken, realm, gocloak.GetClientsParams{ClientID: &clientId})
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return "", err
	}
	return *clients[0].ID, nil
}

func (k *Keycloak) getClientRole(ctx context.Context, accessToken string, realm string, clientUuid string,
	roleName string) (*gocloak.Role, error) {
	role, err := k.client.GetClientRole(context.Background(), accessToken, realm, clientUuid, roleName)
	if err != nil {
		log.Error(ctx, "Getting Client Role is failed", err)
		return nil, err
	}
	return role, nil
}

func (k *Keycloak) addClientRoleToGroup(ctx context.Context, accessToken string, realm string, clientUuid string,
	groupUuid string, role *gocloak.Role) error {
	err := k.client.AddClientRolesToGroup(context.Background(), accessToken, realm, clientUuid, groupUuid, []gocloak.Role{*role})
	if err != nil {
		log.Error(ctx, "Adding Client Role to Group is failed", err)
		return err
	}
	return nil
}

func (k *Keycloak) createClientProtocolMapper(ctx context.Context, accessToken string, realm string,
	id string, mapper gocloak.ProtocolMapperRepresentation) (string, error) {
	id, err := k.client.CreateClientProtocolMapper(context.Background(), accessToken, realm, id, mapper)
	if err != nil {
		log.Error(ctx, "Creating Client Protocol Mapper is failed", err)
		return "", err
	}

	return id, nil
}

func (k *Keycloak) createDefaultClient(ctx context.Context, accessToken string, realm string, clientId string,
	clientSecret string, redirectURIs *[]string) (string, error) {
	id, err := k.client.CreateClient(context.Background(), accessToken, realm, gocloak.Client{
		ClientID:                  gocloak.StringP(clientId),
		DirectAccessGrantsEnabled: gocloak.BoolP(true),
		Enabled:                   gocloak.BoolP(true),
		RedirectURIs:              redirectURIs,
	})

	if err != nil {
		log.Error(ctx, "Creating Client is failed", err)
		return "", err
	}
	client, err := k.client.GetClient(context.Background(), accessToken, realm, id)
	if err != nil {
		log.Error(ctx, "Getting Client is failed", err)
		return "", err
	}
	client.Secret = gocloak.StringP(clientSecret)
	err = k.client.UpdateClient(context.Background(), accessToken, realm, *client)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (k *Keycloak) reflectOrganization(org gocloak.RealmRepresentation) *model.Organization {
	return &model.Organization{
		ID:   *org.ID,
		Name: *org.Realm,
	}
}

func (k *Keycloak) reflectRealmRepresentation(org model.Organization) *gocloak.RealmRepresentation {
	return &gocloak.RealmRepresentation{
		Realm: gocloak.StringP(org.Name),
	}
}

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
	{
		Name:           gocloak.StringP("project-role"),
		Protocol:       gocloak.StringP("openid-connect"),
		ProtocolMapper: gocloak.StringP("oidc-usermodel-client-role-mapper"),
		Config: &map[string]string{
			"access.token.claim":   "true",
			"id.token.claim":       "false",
			"userinfo.token.claim": "false",

			"claim.name":     "project-role",
			"jsonType.label": "String",
			"multivalued":    "true",

			"usermodel.clientRoleMapping.clientId":    "tks",
			"usermodel.clientRoleMapping.role_prefix": "",
		},
	},
}

func defaultRealmSetting(realmId string) gocloak.RealmRepresentation {
	return gocloak.RealmRepresentation{
		Realm:                 gocloak.StringP(realmId),
		Enabled:               gocloak.BoolP(true),
		AccessTokenLifespan:   gocloak.IntP(accessTokenLifespan),
		SsoSessionIdleTimeout: gocloak.IntP(ssoSessionIdleTimeout),
		SsoSessionMaxLifespan: gocloak.IntP(ssoSessionMaxLifespan),
	}
}

func getRefreshTokenExpiredDuration(token *gocloak.JWT) time.Duration {
	return time.Duration(token.RefreshExpiresIn) * time.Second
}
