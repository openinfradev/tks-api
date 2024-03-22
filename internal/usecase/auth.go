package usecase

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"

	"github.com/Nerzal/gocloak/v13"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/mail"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuthUsecase interface {
	Login(ctx context.Context, accountId string, password string, organizationId string) (model.User, error)
	Logout(ctx context.Context, sessionId string, organizationId string) error
	FindId(ctx context.Context, code string, email string, userName string, organizationId string) (string, error)
	FindPassword(ctx context.Context, code string, accountId string, email string, userName string, organizationId string) error
	VerifyIdentity(ctx context.Context, accountId string, email string, userName string, organizationId string) error
	SingleSignIn(ctx context.Context, organizationId, accountId, password string) ([]*http.Cookie, error)
	SingleSignOut(ctx context.Context, organizationId string) (string, []*http.Cookie, error)
	VerifyToken(ctx context.Context, token string) (bool, error)
	UpdateExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) error
}

const (
	passwordLength                  = 8
	KEYCLOAK_IDENTITY_COOKIE        = "KEYCLOAK_IDENTITY"
	KEYCLOAK_IDENTITY_LEGACY_COOKIE = "KEYCLOAK_IDENTITY_LEGACY"
)

type AuthUsecase struct {
	kc                     keycloak.IKeycloak
	userRepository         repository.IUserRepository
	authRepository         repository.IAuthRepository
	clusterRepository      repository.IClusterRepository
	appgroupRepository     repository.IAppGroupRepository
	organizationRepository repository.IOrganizationRepository
}

func NewAuthUsecase(r repository.Repository, kc keycloak.IKeycloak) IAuthUsecase {
	return &AuthUsecase{
		kc:                     kc,
		userRepository:         r.User,
		authRepository:         r.Auth,
		clusterRepository:      r.Cluster,
		appgroupRepository:     r.AppGroup,
		organizationRepository: r.Organization,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, accountId string, password string, organizationId string) (model.User, error) {
	// Authentication with DB
	user, err := u.userRepository.Get(ctx, accountId, organizationId)
	if err != nil {
		return model.User{}, httpErrors.NewBadRequestError(err, "A_INVALID_ID", "")
	}

	var accountToken *model.User
	accountToken, err = u.kc.Login(ctx, accountId, password, organizationId)
	if err != nil {
		apiErr, ok := err.(*gocloak.APIError)
		if ok {
			if apiErr.Code == 401 {
				return model.User{}, httpErrors.NewBadRequestError(fmt.Errorf("Mismatch password"), "A_INVALID_PASSWORD", "")
			}
		}
		return model.User{}, httpErrors.NewInternalServerError(err, "", "")
	}

	// Insert token
	user.Token = accountToken.Token

	if !(organizationId == "master" && accountId == "admin") {
		user.PasswordExpired = helper.IsDurationExpired(user.PasswordUpdatedAt, internal.PasswordExpiredDuration)
	}

	return user, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionId string, organizationName string) error {
	// [TODO] refresh token 을 추가하고, session timeout 을 줄이는 방향으로 고려할 것
	err := u.kc.Logout(ctx, sessionId, organizationName)
	if err != nil {
		return err
	}
	return nil
}

func (u *AuthUsecase) FindId(ctx context.Context, code string, email string, userName string, organizationId string) (string, error) {
	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.NameFilter(userName), u.userRepository.EmailFilter(email))
	if err != nil && users == nil {
		return "", httpErrors.NewBadRequestError(err, "A_INVALID_ID", "")
	}
	if err != nil {
		return "", httpErrors.NewInternalServerError(err, "", "")
	}
	emailCode, err := u.authRepository.GetEmailCode(ctx, (*users)[0].ID)
	if err != nil {
		return "", httpErrors.NewInternalServerError(err, "", "")
	}
	if !u.isExpiredEmailCode(emailCode) {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("expired code"), "A_EXPIRED_CODE", "")
	}
	if emailCode.Code != code {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("invalid code"), "A_INVALID_CODE", "")
	}
	if err := u.authRepository.DeleteEmailCode(ctx, (*users)[0].ID); err != nil {
		return "", httpErrors.NewInternalServerError(err, "", "")
	}

	return (*users)[0].AccountId, nil
}

func (u *AuthUsecase) FindPassword(ctx context.Context, code string, accountId string, email string, userName string, organizationId string) error {
	users, err := u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
		u.userRepository.AccountIdFilter(accountId), u.userRepository.NameFilter(userName),
		u.userRepository.EmailFilter(email))
	if err != nil && users == nil {
		return httpErrors.NewBadRequestError(err, "A_INVALID_ID", "")
	}
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	user := (*users)[0]
	emailCode, err := u.authRepository.GetEmailCode(ctx, user.ID)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	if !u.isExpiredEmailCode(emailCode) {
		return httpErrors.NewBadRequestError(fmt.Errorf("expired code"), "A_EXPIRED_CODE", "")
	}
	if emailCode.Code != code {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid code"), "A_INVALID_CODE", "")
	}
	randomPassword := helper.GenerateRandomString(passwordLength)

	originUser, err := u.kc.GetUser(ctx, organizationId, accountId)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	originUser.Credentials = &[]gocloak.CredentialRepresentation{
		{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(randomPassword),
			Temporary: gocloak.BoolP(false),
		},
	}
	if err = u.kc.UpdateUser(ctx, organizationId, originUser); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	if err = u.userRepository.UpdatePasswordAt(ctx, user.ID, organizationId, true); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	message, err := mail.MakeTemporaryPasswordMessage(ctx, email, organizationId, accountId, randomPassword)
	if err != nil {
		log.Errorf(ctx, "mail.MakeVerityIdentityMessage error. %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(ctx); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	if err = u.authRepository.DeleteEmailCode(ctx, user.ID); err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (u *AuthUsecase) VerifyIdentity(ctx context.Context, accountId string, email string, userName string, organizationId string) error {
	var users *[]model.User
	var err error

	if accountId == "" {
		users, err = u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
			u.userRepository.NameFilter(userName), u.userRepository.EmailFilter(email))
	} else {
		users, err = u.userRepository.List(ctx, u.userRepository.OrganizationFilter(organizationId),
			u.userRepository.AccountIdFilter(accountId), u.userRepository.NameFilter(userName),
			u.userRepository.EmailFilter(email))
	}
	if err != nil && users == nil {
		return httpErrors.NewBadRequestError(err, "A_INVALID_ID", "")
	}
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}

	code, err := helper.GenerateEmailCode(ctx)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	_, err = u.authRepository.GetEmailCode(ctx, (*users)[0].ID)
	if err != nil {
		if err := u.authRepository.CreateEmailCode(ctx, (*users)[0].ID, code); err != nil {
			return httpErrors.NewInternalServerError(err, "", "")
		}
	} else {
		if err := u.authRepository.UpdateEmailCode(ctx, (*users)[0].ID, code); err != nil {
			return httpErrors.NewInternalServerError(err, "", "")
		}
	}

	message, err := mail.MakeVerityIdentityMessage(ctx, email, code)
	if err != nil {
		log.Errorf(ctx, "mail.MakeVerityIdentityMessage error. %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	mailer := mail.New(message)

	if err := mailer.SendMail(ctx); err != nil {
		log.Errorf(ctx, "mailer.SendMail error. %v", err)
		return httpErrors.NewInternalServerError(err, "", "")
	}

	return nil
}

func (u *AuthUsecase) SingleSignIn(ctx context.Context, organizationId, accountId, password string) ([]*http.Cookie, error) {
	cookies, err := makingCookie(ctx, organizationId, accountId, password)
	if err != nil {
		return nil, err
	}
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no cookie generated")
	}

	return cookies, nil
}

func (u *AuthUsecase) SingleSignOut(ctx context.Context, organizationId string) (string, []*http.Cookie, error) {
	var redirectUrl string

	organization, err := u.organizationRepository.Get(ctx, organizationId)
	if err != nil {
		return "", nil, err
	}

	appGroupsInPrimaryCluster, err := u.appgroupRepository.Fetch(ctx, domain.ClusterId(organization.PrimaryClusterId), nil)
	if err != nil {
		return "", nil, err
	}

	for _, appGroup := range appGroupsInPrimaryCluster {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appgroupRepository.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
			if err != nil {
				return "", nil, err
			}
			if len(applications) > 0 {
				redirectUrl = "http://" + applications[0].Endpoint + "/logout"
			}
		}
	}

	// cookies to be deleted
	cookies := []*http.Cookie{
		{
			Name:     KEYCLOAK_IDENTITY_COOKIE,
			MaxAge:   -1,
			Expires:  time.Now().AddDate(0, 0, -1),
			Path:     "/auth/realms/" + organizationId + "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		},
		{
			Name:     KEYCLOAK_IDENTITY_LEGACY_COOKIE,
			MaxAge:   -1,
			Expires:  time.Now().AddDate(0, 0, -1),
			Path:     "/auth/realms/" + organizationId + "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		},
	}

	return redirectUrl, cookies, nil
}

func (u *AuthUsecase) VerifyToken(ctx context.Context, token string) (bool, error) {
	parsedToken, err := helper.StringToTokenWithoutVerification(token)
	if err != nil {
		return false, err
	}
	claims, err := helper.RetrieveClaims(parsedToken)
	if err != nil {
		return false, err
	}

	org, ok := claims["organization"].(string)
	if !ok {
		return false, fmt.Errorf("organization is not found in token")
	}

	isActive, err := u.kc.VerifyAccessToken(ctx, token, org)
	if err != nil {
		return false, err
	}
	if !isActive {
		return false, nil
	}

	return true, nil
}

func (u *AuthUsecase) UpdateExpiredTimeOnToken(ctx context.Context, organizationId string, userId string) error {
	return u.authRepository.UpdateExpiredTimeOnToken(ctx, organizationId, userId)
}

func (u *AuthUsecase) isExpiredEmailCode(code model.CacheEmailCode) bool {
	return !helper.IsDurationExpired(code.UpdatedAt, internal.EmailCodeExpireTime)
}

func extractFormAction(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == "kc-form-login" {
					for _, a := range n.Attr {
						if a.Key == "action" {
							return a.Val
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := f(c); result != "" {
				return result
			}
		}
		return ""
	}

	return f(doc), nil
}

func makingCookie(ctx context.Context, organizationId, userName, password string) ([]*http.Cookie, error) {
	stateCode, err := genStateString()
	if err != nil {
		return nil, err
	}
	baseUrl := viper.GetString("keycloak-address") + "/realms/" + organizationId + "/protocol/openid-connect"
	var oauth2Config = &oauth2.Config{
		ClientID:     keycloak.DefaultClientID,
		ClientSecret: viper.GetString("keycloak-client-secret"),
		RedirectURL:  viper.GetString("external-address") + internal.API_PREFIX + internal.API_VERSION + "/auth/callback",
		Scopes:       []string{"openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseUrl + "/auth",
			TokenURL: baseUrl + "/token",
		},
	}

	authCodeUrl := oauth2Config.AuthCodeURL(stateCode, oauth2.AccessTypeOnline)
	// skip tls check
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", authCodeUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "text/html")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(ctx, "Error while creating new request: %v", err)
		return nil, err
	}
	cookies := resp.Cookies()
	if len(cookies) < 1 {
		return nil, fmt.Errorf("no cookie found")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	htmlContent := string(body)

	s, err := extractFormAction(htmlContent)
	if err != nil {
		log.Errorf(ctx, "Error while creating new request: %v", err)
		return nil, err
	}

	data := url.Values{}
	data.Set("username", userName)
	data.Set("password", password)

	req, err = http.NewRequest("POST", s, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	cookies = resp.Cookies()
	var targetCookies []*http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == KEYCLOAK_IDENTITY_COOKIE || cookie.Name == KEYCLOAK_IDENTITY_LEGACY_COOKIE {
			targetCookies = append(targetCookies, cookie)
		}
	}

	return targetCookies, nil
}

func genStateString() (string, error) {
	rnd := make([]byte, 32)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rnd), nil
}
