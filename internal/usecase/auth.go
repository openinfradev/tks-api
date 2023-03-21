package usecase

import (
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IAuthUsecase interface {
	Login(accountId string, password string, organizationName string) (domain.User, error)
	Register(accountId string, password string, name string, organizationName string, role string, token string) (domain.User, error)
	FetchRoles() (out []domain.Role, err error)
	//AuthenticateToken(organization string, accessToken string) (*authenticator.Response, bool, error)
}

type AuthUsecase struct {
	kc   keycloak.IKeycloak
	repo repository.IAuthRepository
}

func NewAuthUsecase(r repository.IAuthRepository, kc keycloak.IKeycloak) IAuthUsecase {
	return &AuthUsecase{
		repo: r,
		kc:   kc,
	}
}

func (r *AuthUsecase) Login(accountId string, password string, organizationName string) (domain.User, error) {
	user, err := r.repo.GetUserByAccountId(accountId)
	if err != nil {
		return domain.User{}, err
	}

	// Authentication with Keycloak
	accountToken, err := r.kc.GetAccessTokenByIdPassword(accountId, password, organizationName)
	if err != nil {
		return domain.User{}, err
	}

	// Insert token
	user.Token = accountToken.Token

	// Authentication with DB
	//
	//if !helper.CheckPasswordHash(user.Password, password) {
	//	log.Debug(user.Password)
	//	log.Debug(password)
	//	return domain.User{}, fmt.Errorf("Invalid password")
	//}
	//
	//user.Token, err = helper.CreateJWT(accountId, user.ID)
	//if err != nil {
	//	return domain.User{}, fmt.Errorf("failed to create token")
	//}

	return user, nil
}

func (r *AuthUsecase) Register(accountId string, password string, name string, organizationName string, role string, accessToken string) (domain.User, error) {
	// Validation check
	user, err := r.kc.GetUser(organizationName, accountId, accessToken)
	if err != nil {
		return domain.User{}, err
	}
	if user != nil {
		return domain.User{}, fmt.Errorf("Already existed user. %s", accountId)
	}
	_, err = r.repo.GetUserByAccountId(accountId)
	if err == nil {
		return domain.User{}, fmt.Errorf("Already existed user. %s", accountId)
	}

	// Create user in keycloak
	groups := []string{fmt.Sprintf("%s@%s", role, organizationName)}
	err = r.kc.CreateUser(organizationName, &gocloak.User{
		Username: gocloak.StringP(accountId),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(password),
				Temporary: gocloak.BoolP(false),
			},
		},
		Groups: &groups,
	}, accessToken)
	if err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	resUser, err := r.repo.Create(accountId, hashedPassword, name)
	if err != nil {
		return domain.User{}, err
	}

	// [TODO] 임시로 tks-admin 으로 세팅한다.
	err = r.repo.AssignRole(accountId, "tks-admin")
	if err != nil {
		return domain.User{}, err
	}

	return resUser, nil
}

func (u *AuthUsecase) FetchRoles() (out []domain.Role, err error) {
	roles, err := u.repo.FetchRoles()
	if err != nil {
		return nil, err
	}
	return roles, nil
}

//func (r *AuthUsecase) AuthenticateToken(organization string, accessToken string) (*authenticator.Response, bool, error) {
//	if err := r.kc.VerifyAccessToken(accessToken, organization); err != nil {
//		return nil, false, err
//	}
//	jwtToken, mapClaims, err := r.kc.ParseAccessToken(accessToken, organization)
//	if jwtToken == nil || mapClaims == nil || mapClaims.Valid() != nil {
//		return nil, false, err
//	}
//	roleProjectMapping := make(map[string]string)
//	for _, role := range jwtToken.Claims.(jwt.MapClaims)["tks-role"].([]interface{}) {
//		slice := strings.Split(role.(string), "@")
//		if len(slice) != 2 {
//			return nil, false, nil
//		}
//		// key is projectName and value is roleName
//		roleProjectMapping[slice[1]] = slice[0]
//	}
//	log.Info("Valid Authentication")
//
//	return &authenticator.Response{
//		User: &user.DefaultInfo{
//			Organization:       jwtToken.Claims.(jwt.MapClaims)["organization"].(string),
//			RoleProjectMapping: roleProjectMapping,
//		},
//	}, true, nil
//}
