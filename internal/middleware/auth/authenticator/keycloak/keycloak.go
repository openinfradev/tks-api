package keycloak

import (
	"fmt"
	"net/http"
	"strings"

	jwtWithouKey "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/auth/user"
	"github.com/openinfradev/tks-api/pkg/log"
)

type keycloakAuthenticator struct {
	kc keycloak.IKeycloak
}

func NewKeycloakAuthenticator(kc keycloak.IKeycloak) *keycloakAuthenticator {
	return &keycloakAuthenticator{
		kc: kc,
	}
}

func (a *keycloakAuthenticator) AuthenticateRequest(r *http.Request) (*authenticator.Response, bool, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return nil, false, fmt.Errorf("authorizer header is invalid")
	}
	parts := strings.SplitN(authHeader, " ", 3)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, false, fmt.Errorf("authorizer header is invalid")
	}

	token := parts[1]

	if len(token) == 0 {
		// The space before the token case
		if len(parts) == 3 {
			log.Warn("the provided Authorization header contains extra space before the bearer token, and is ignored")
		}
		return nil, false, fmt.Errorf("token is empty")
	}

	return a.AuthenticateToken(r, token)
}

func (a *keycloakAuthenticator) AuthenticateToken(r *http.Request, token string) (*authenticator.Response, bool, error) {
	parsedToken, _, err := new(jwtWithouKey.Parser).ParseUnverified(token, jwtWithouKey.MapClaims{})
	if err != nil {
		return nil, false, err
	}

	if parsedToken.Method.Alg() != "RS256" {
		return nil, false, fmt.Errorf("invalid token")
	}

	if parsedToken.Claims.Valid() != nil {
		return nil, false, fmt.Errorf("invalid token")
	}

	organizationId, ok := parsedToken.Claims.(jwtWithouKey.MapClaims)["organization"].(string)
	if !ok {
		return nil, false, fmt.Errorf("organization is not found in token")
	}

	if err := a.kc.VerifyAccessToken(token, organizationId); err != nil {
		log.Errorf("failed to verify access token: %v", err)
		return nil, false, err
	}

	roleProjectMapping := make(map[string]string)
	for _, role := range parsedToken.Claims.(jwtWithouKey.MapClaims)["tks-role"].([]interface{}) {
		slice := strings.Split(role.(string), "@")
		if len(slice) != 2 {
			log.Errorf("invalid tks-role format: %v", role)

			return nil, false, fmt.Errorf("invalid tks-role format")
		}
		// key is projectName and value is roleName
		roleProjectMapping[slice[1]] = slice[0]
	}
	userAccountId, err := uuid.Parse(parsedToken.Claims.(jwtWithouKey.MapClaims)["sub"].(string))
	if err != nil {
		log.Errorf("failed to verify access token: %v", err)

		return nil, false, err
	}
	requestSessionId, ok := parsedToken.Claims.(jwtWithouKey.MapClaims)["sid"].(string)
	if !ok {
		return nil, false, fmt.Errorf("session id is not found in token")
	}

	sessionIds, err := a.kc.GetSessions(userAccountId.String(), organizationId)
	if err != nil {
		log.Errorf("failed to get sessions: %v", err)

		return nil, false, err
	}
	if len(*sessionIds) == 0 {
		return nil, false, fmt.Errorf("invalid session")
	}
	var matched bool = false
	for _, id := range *sessionIds {
		if id == requestSessionId {
			matched = true
			break
		}
	}
	if !matched {
		return nil, false, fmt.Errorf("invalid session")
	}

	userInfo := &user.DefaultInfo{
		UserId:             userAccountId,
		OrganizationId:     organizationId,
		RoleProjectMapping: roleProjectMapping,
	}
	//r = r.WithContext(request.WithToken(r.Context(), token))
	*r = *(r.WithContext(request.WithToken(r.Context(), token)))

	return &authenticator.Response{User: userInfo}, true, nil
}
