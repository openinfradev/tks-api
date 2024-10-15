package keycloak

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"net/http"
	"strings"

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
			log.Warn(r.Context(), "the provided Authorization header contains extra space before the bearer token, and is ignored")
		}
		return nil, false, fmt.Errorf("token is empty")
	}

	return a.AuthenticateToken(r, token)
}

func (a *keycloakAuthenticator) AuthenticateToken(r *http.Request, token string) (*authenticator.Response, bool, error) {
	parsedToken, err := helper.StringToTokenWithoutVerification(token)
	if err != nil {
		return nil, false, httpErrors.NewUnauthorizedError(err, "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	claims, err := helper.RetrieveClaims(parsedToken)
	if err != nil {
		return nil, false, httpErrors.NewUnauthorizedError(err, "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	organizationId, ok := claims["organization"].(string)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("organization is not found in token"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	isActive, err := a.kc.VerifyAccessToken(r.Context(), token, organizationId)
	if err != nil {
		log.Errorf(r.Context(), "failed to verify access token: %v", err)
		return nil, false, httpErrors.NewUnauthorizedError(err, "C_INTERNAL_ERROR", "")
	}
	if !isActive {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("token is deactivated"), "A_EXPIRED_TOKEN", "토큰이 만료되었습니다.")
	}

	// tks role extraction
	roleOrganizationMapping := make(map[string]string)
	if roles, ok := claims["tks-role"]; !ok {
		log.Errorf(r.Context(), "tks-role is not found in token")

		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("tks-role is not found in token"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	} else {
		for _, role := range roles.([]interface{}) {
			slice := strings.Split(role.(string), "@")
			if len(slice) != 2 {
				log.Errorf(r.Context(), "invalid tks-role format: %v", role)

				return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid tks-role format"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
			}
			// key is projectName and value is roleName
			roleOrganizationMapping[slice[1]] = slice[0]
		}

	}
	// project role extraction
	projectIds := make([]string, 0)
	roleProjectMapping := make(map[string]string)
	if roles, ok := claims["project-role"]; ok {
		for _, role := range roles.([]interface{}) {
			slice := strings.Split(role.(string), "@")
			if len(slice) != 2 {
				log.Errorf(r.Context(), "invalid project-role format: %v", role)

				return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid project-role format"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
			}
			// key is projectId and value is roleName
			roleProjectMapping[slice[1]] = slice[0]
			projectIds = append(projectIds, slice[1])
		}
	}

	userId, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		log.Errorf(r.Context(), "failed to verify access token: %v", err)

		return nil, false, httpErrors.NewUnauthorizedError(err, "C_INTERNAL_ERROR", "")
	}
	requestSessionId, ok := claims["sid"].(string)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("session id is not found in token"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	userAccountId, ok := claims["preferred_username"].(string)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("preferred_username is not found in token"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	userInfo := &user.DefaultInfo{
		UserId:                  userId,
		AccountId:               userAccountId,
		OrganizationId:          organizationId,
		ProjectIds:              projectIds,
		RoleOrganizationMapping: roleOrganizationMapping,
		RoleProjectMapping:      roleProjectMapping,
	}
	//r = r.WithContext(request.WithToken(r.Context(), token))
	*r = *(r.WithContext(request.WithToken(r.Context(), token)))

	*r = *(r.WithContext(request.WithSession(r.Context(), requestSessionId)))

	return &authenticator.Response{User: userInfo}, true, nil
}
