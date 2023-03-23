package helper

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func CreateJWT(accountId string, uId string, organizationId string) (string, error) {
	signingKey := []byte(viper.GetString("jwt-secret"))

	aToken := jwt.New(jwt.SigningMethodHS256)
	claims := aToken.Claims.(jwt.MapClaims)
	claims["AccountId"] = accountId
	claims["ID"] = uId
	claims["OrganizationId"] = organizationId
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tk, err := aToken.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return tk, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	signingKey := []byte(viper.GetString("jwt-secret"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, err
}
