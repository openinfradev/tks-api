package keycloak

import (
	"fmt"
	jwtWithouKey "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/auth/user"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
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
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("authorization header is invalid"))
	}
	parts := strings.SplitN(authHeader, " ", 3)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("authorization header is invalid"))
	}

	token := parts[1]

	if len(token) == 0 {
		// The space before the token case
		if len(parts) == 3 {
			log.Warn("the provided Authorization header contains extra space before the bearer token, and is ignored")
		}
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("authorization header is invalid"))
	}

	return a.AuthenticateToken(r, token)
}

func (a *keycloakAuthenticator) AuthenticateToken(r *http.Request, token string) (*authenticator.Response, bool, error) {
	parsedToken, _, err := new(jwtWithouKey.Parser).ParseUnverified(token, jwtWithouKey.MapClaims{})
	if err != nil {
		return nil, false, err
	}
	organizationId, ok := parsedToken.Claims.(jwtWithouKey.MapClaims)["organization"].(string)
	if !ok {
		return nil, false, fmt.Errorf("organization is not found in token")
	}
	if err := a.kc.VerifyAccessToken(token, organizationId); err != nil {
		return nil, false, err
	}

	roleProjectMapping := make(map[string]string)
	for _, role := range parsedToken.Claims.(jwtWithouKey.MapClaims)["tks-role"].([]interface{}) {
		slice := strings.Split(role.(string), "@")
		if len(slice) != 2 {
			return nil, false, fmt.Errorf("invalid tks-role format")
		}
		// key is projectName and value is roleName
		roleProjectMapping[slice[1]] = slice[0]
	}
	userAccountId, err := uuid.Parse(parsedToken.Claims.(jwtWithouKey.MapClaims)["sub"].(string))
	if err != nil {
		return nil, false, err
	}

	userInfo := &user.DefaultInfo{
		UserId:             userAccountId,
		OrganizationId:     organizationId,
		RoleProjectMapping: roleProjectMapping,
	}
	*r = *(r.WithContext(request.WithToken(r.Context(), token)))

	return &authenticator.Response{User: userInfo}, true, nil
}
