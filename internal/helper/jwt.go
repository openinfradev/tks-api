package helper

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

func CreateJWT(accountId string, uId string, organizationId string) (string, error) {
	signingKey := []byte(viper.GetString("jwt-secret"))

	aToken := jwt.New(jwt.SigningMethodHS256)
	claims, ok := aToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil
	}
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

func StringToTokenWithoutVerification(tokenString string) (*jwt.Token, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}
func RetrieveClaims(token *jwt.Token) (map[string]interface{}, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
