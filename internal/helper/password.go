package helper

import (
	"crypto/rand"
	"golang.org/x/crypto/bcrypt"
	"math/big"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hashVal, userPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashVal), []byte(userPw))
	if err != nil {
		return false
	} else {
		return true
	}
}

func GenerateRandomString(length int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = randomRune(letters)
	}
	return string(b)
}

func randomRune(chars []rune) rune {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		panic(err)
	}
	return chars[n.Int64()]
}
