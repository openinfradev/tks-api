package helper_test

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"testing"
)

func TestPasswordGeneration(t *testing.T) {
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
