package helper_test

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "admin"
	hashedPassword, err := helper.HashPassword(password)
	fmt.Println(hashedPassword)
	if err != nil {
		t.Errorf("HashPassword() error = %v", err)
		return
	}
	if !helper.CheckPasswordHash(hashedPassword, password) {
		t.Errorf("CheckPasswordHash() error")
		return
	}
}

func TestPasswordGeneration(t *testing.T) {
	length := 8
	sameCount := 0
	for i := 0; i < 10000000; i++ {
		firstGenPassword := helper.GenerateRandomString(length)
		secondGenPassword := helper.GenerateRandomString(length)
		if firstGenPassword == secondGenPassword {
			fmt.Printf("Index: %d. It's same %s\n", i, firstGenPassword)
			sameCount++
		}
	}
	fmt.Println("Same count: ", sameCount)
	fmt.Println("Same ratio: ", float64(sameCount)/100000)
}
